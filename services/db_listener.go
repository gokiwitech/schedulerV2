package services

import (
	"math"
	"schedulerV2/config"
	"schedulerV2/models"
	"schedulerV2/repositories"
	"schedulerV2/zkclient"
	"strconv"
	"sync"
	"time"

	"gorm.io/gorm"
)

var messageQueueRepository *repositories.MessageQueueRepository
var thresholdRepository *repositories.ServiceThresholdRepository
var lg = config.GetLogger(true)

func InitServices() {
	messageQueueRepository = repositories.NewMessageQueueRepository()
	thresholdRepository = repositories.NewServiceThresholdRepository()
}

func StartSchedulers() {
	tickerDLQ := time.NewTicker(2300 * time.Millisecond)
	tickerProcess := time.NewTicker(1000 * time.Millisecond)

	go func() {
		for {
			select {
			case <-tickerDLQ.C:
				go updateScheduledDLQMessageStatus()
			case <-tickerProcess.C:
				go scanAndProcessScheduledMessages()
				go scanAndProcessCronMessages()
			}
		}
	}()

}

func updateScheduledDLQMessageStatus() {
	lg.Info().Msg("Checking for DLQ status for messages whose retry count is 20...")

	db, err := config.GetDBConnection()
	if err != nil {
		lg.Error().Msgf("Error getting database connection: %v", err)
		return
	}
	messages, err := messageQueueRepository.FindByStatusAndRetryCountAndIsDLQ(db, string(models.PENDING), string(models.SCHEDULED), false, models.AppConfig.DlqMessageLimit)
	if err != nil {
		lg.Error().Msgf("Error fetching messages: %v", err)
		return
	}

	for _, message := range messages {

		// Acquire the lock with a distinct DLQ identifier before starting the transaction
		lockName := zkclient.LockName + "DLQ" + strconv.Itoa(int(message.ID))
		lock := zkclient.NewDistributedLock(config.ZkConn, zkclient.LockBasePath, lockName)
		acquired, err := lock.Acquire()
		if err != nil {
			lg.Error().Msgf("Error acquiring DLQ lock for message ID %d: %v", message.ID, err)
			continue
		}
		if !acquired {
			// DLQ Lock not acquired, skip this message
			continue
		}

		// Ensure the lock is released after the transaction is done
		defer func() {
			if err := lock.Release(); err != nil {
				lg.Error().Msgf("Error releasing DLQ lock for message ID %d: %v", message.ID, err)
			}
		}()

		// Start a transaction
		tx := db.Begin()
		if tx.Error != nil {
			lg.Error().Msgf("Error starting transaction: %v", tx.Error)
			continue
		}

		dlqMessage := models.DlqMessageQueue{MessageID: message.ID, IsProcessed: false}
		message.IsDLQ = true

		// Save DLQ message within the transaction
		if err := tx.Create(&dlqMessage).Error; err != nil {
			tx.Rollback()
			lg.Error().Msgf("Error - %v saving DLQ message: %v", err, message)
			continue
		}

		// Save message status within the transaction
		if err := tx.Save(&message).Error; err != nil {
			tx.Rollback()
			lg.Error().Msgf("Error - %v updating message status: %v", err, message)
			continue
		}

		// Commit the transaction
		if err := tx.Commit().Error; err != nil {
			lg.Error().Msgf("Transaction commit failed: %v", err)
			continue
		}

		lg.Info().Msgf("Message with id %d marked as DLQ and moved to DLQ table", message.ID)
	}
}

func scanAndProcessScheduledMessages() {
	lg.Info().Msg("Scanning for pending messages and processing...")

	db, err := config.GetDBConnection()
	if err != nil {
		lg.Error().Msgf("Error getting database connection: %v", err)
		return
	}
	nowPlusOneSecond := time.Now().Unix() + 1
	messages, err := messageQueueRepository.FindByStatusAndNextRetryAndRetryCountAndIsDLQ(db, string(models.PENDING), string(models.SCHEDULED), false, models.AppConfig.DlqMessageLimit, nowPlusOneSecond)
	if err != nil {
		lg.Error().Msgf("Error fetching messages: %v", err)
		return
	}
	var wg sync.WaitGroup

	for _, message := range messages {
		wg.Add(1)
		go func(msg models.MessageQueue) {
			// Decrement the counter when the go routine completes
			defer wg.Done()

			currentTime := time.Now().Unix()
			threshold, err := thresholdRepository.FindByServiceName(db, msg.ServiceName, currentTime)
			if err != nil || (threshold != nil && !thresholdRepository.IsWithinThreshold(threshold)) {
				msg.Status = models.DEAD
				if saveErr := messageQueueRepository.Save(db, &msg); saveErr != nil {
					lg.Error().Msgf("Failed to mark message ID %d as DEAD: %v", msg.ID, saveErr)
				}

				if err != nil && err != gorm.ErrRecordNotFound {
					lg.Error().Msgf("Error checking service threshold for message ID %d: %v", msg.ID, err)
				} else {
					lg.Info().Msgf("Message ID %d marked as DEAD - outside service threshold", msg.ID)
				}
				return
			}

			lock := zkclient.NewDistributedLock(config.ZkConn, zkclient.LockBasePath, zkclient.LockName+strconv.Itoa(int(msg.ID)))
			acquired, err := lock.Acquire()
			if err != nil {
				lg.Error().Msgf("Error acquiring lock for message ID %d: %v", msg.ID, err)
				return
			}
			if !acquired {
				// Lock not acquired, another process is already processing this message
				return
			}

			if msg.RetryCount >= models.AppConfig.DlqMessageLimit {
				msg.Status = models.COMPLETED
				messageQueueRepository.Save(db, &msg)
				return
			}

			defer lock.Release()

			// Update the message status to IN_PROGRESS in the database
			if err := setMessageStatusInProgress(db, &msg); err != nil {
				lg.Error().Msgf("Failed to set IN-PROGRESS status for message ID %d: %v", msg.ID, err)
				return
			}

			// Proceed with processing the message
			if err := processScheduledMessage(&msg); err != nil {
				lg.Error().Msgf("Error processing message ID %d: %v", msg.ID, err)
				message.Status = models.PENDING
				messageQueueRepository.Save(db, &msg)
			}
		}(message)
	}
	wg.Wait()
}

func scanAndProcessCronMessages() {
	lg.Info().Msg("Scanning & Processing cron messages...")

	db, err := config.GetDBConnection()
	if err != nil {
		lg.Error().Msgf("Error getting database connection: %v", err)
		return
	}
	// Fetch cron messages
	nowPlusOneSecond := time.Now().Unix() + 1
	cronMessages, err := messageQueueRepository.FindByStatusAndNextRetryAndRetryCountAndIsDLQ(db, string(models.PENDING), string(models.CRON), false, int(math.Inf(1)), nowPlusOneSecond)
	if err != nil {
		lg.Error().Msgf("Error fetching cron messages: %v", err)
		return
	}

	var wg sync.WaitGroup
	for _, message := range cronMessages {
		wg.Add(1)
		go func(msg models.MessageQueue) {
			defer wg.Done()

			if msg.Count != -1 && msg.Count < msg.RetryCount {
				msg.Status = models.COMPLETED
				messageQueueRepository.Save(db, &msg)
				return
			}

			// Calculate the next run time from the next retry time
			nextRetryTime := time.Unix(msg.NextRetry, 0).UTC()
			if time.Now().UTC().Before(nextRetryTime) {
				return
			}

			// Acquire distributed lock
			lock := zkclient.NewDistributedLock(config.ZkConn, zkclient.LockBasePath, zkclient.LockName+strconv.Itoa(int(msg.ID)))
			acquired, err := lock.Acquire()
			if err != nil {
				lg.Error().Msgf("Error acquiring lock for message ID %d: %v", msg.ID, err)
				return
			}

			if !acquired {
				// Lock not acquired, another process is already processing this message
				return
			}
			defer lock.Release()

			// Update the message status to IN_PROGRESS in the database
			if err := setMessageStatusInProgress(db, &msg); err != nil {
				lg.Error().Msgf("Failed to set IN-PROGRESS status for message ID %d: %v", msg.ID, err)
				return
			}

			// Time to process the message
			lg.Info().Msgf("Processing message ID %d", msg.ID)
			if err := processCronMessage(&msg); err != nil {
				message.Status = models.PENDING
				messageQueueRepository.Save(db, &msg)
				lg.Error().Msgf("Error processing cron message ID %d: %v", msg.ID, err)
			}
		}(message)
	}
	wg.Wait()
}
