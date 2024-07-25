package services

import (
	"log"
	"schedulerV2/config"
	"schedulerV2/models"
	"schedulerV2/repositories"
	"schedulerV2/zkclient"
	"strconv"
	"sync"
	"time"

	"github.com/robfig/cron"
)

var messageQueueRepository *repositories.MessageQueueRepository

func InitServices() {
	messageQueueRepository = repositories.NewMessageQueueRepository()
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
				go scanAndProcessConditionalMessages()
			}
		}
	}()

}

func updateScheduledDLQMessageStatus() {
	log.Println("Checking for DLQ status for messages whose retry count is 20...")

	db, err := config.GetDBConnection()
	if err != nil {
		log.Println("Error getting database connection:", err)
		return
	}
	messages, err := messageQueueRepository.FindByStatusAndRetryCountAndIsDLQ(db, string(models.PENDING), string(models.SCHEDULED), false, models.AppConfig.DlqMessageLimit)
	if err != nil {
		log.Println("Error fetching messages:", err)
		return
	}

	for _, message := range messages {

		// Acquire the lock with a distinct DLQ identifier before starting the transaction
		lockName := zkclient.LockName + "DLQ" + strconv.Itoa(int(message.ID))
		lock := zkclient.NewDistributedLock(config.ZkConn, zkclient.LockBasePath, lockName)
		acquired, err := lock.Acquire()
		if err != nil {
			log.Printf("Error acquiring DLQ lock for message ID %d: %v", message.ID, err)
			continue
		}
		if !acquired {
			// DLQ Lock not acquired, skip this message
			continue
		}

		// Ensure the lock is released after the transaction is done
		defer func() {
			if err := lock.Release(); err != nil {
				log.Printf("Error releasing DLQ lock for message ID %d: %v", message.ID, err)
			}
		}()

		// Start a transaction
		tx := db.Begin()
		if tx.Error != nil {
			log.Println("Error starting transaction:", tx.Error)
			continue
		}

		dlqMessage := models.DlqMessageQueue{MessageID: message.ID, IsProcessed: false}
		message.IsDLQ = true

		// Save DLQ message within the transaction
		if err := tx.Create(&dlqMessage).Error; err != nil {
			tx.Rollback()
			log.Println("Error saving DLQ message:", err, message)
			continue
		}

		// Save message status within the transaction
		if err := tx.Save(&message).Error; err != nil {
			tx.Rollback()
			log.Println("Error updating message status:", err, message)
			continue
		}

		// Commit the transaction
		if err := tx.Commit().Error; err != nil {
			log.Println("Transaction commit failed:", err)
			continue
		}

		log.Printf("Message with id %d marked as DLQ and moved to DLQ table", message.ID)
	}
}

func scanAndProcessScheduledMessages() {
	log.Println("Scanning for pending messages and processing...")

	db, err := config.GetDBConnection()
	if err != nil {
		log.Println("Error getting database connection:", err)
		return
	}
	nowPlusOneSecond := time.Now().Unix() + 1
	messages, err := messageQueueRepository.FindByStatusAndNextRetryAndRetryCountAndIsDLQ(db, string(models.PENDING), string(models.SCHEDULED), false, models.AppConfig.DlqMessageLimit, nowPlusOneSecond)
	if err != nil {
		log.Println("Error fetching messages:", err)
		return
	}
	var wg sync.WaitGroup

	for _, message := range messages {
		wg.Add(1)
		go func(msg models.MessageQueue) {
			// Decrement the counter when the go routine completes
			defer wg.Done()
			lock := zkclient.NewDistributedLock(config.ZkConn, zkclient.LockBasePath, zkclient.LockName+strconv.Itoa(int(msg.ID)))
			acquired, err := lock.Acquire()
			if err != nil {
				log.Printf("Error acquiring lock for message ID %d: %v", msg.ID, err)
				return
			}
			if !acquired {
				// Lock not acquired, another process is already processing this message
				return
			}

			defer lock.Release()

			// Update the message status to IN_PROGRESS in the database
			if err := setMessageStatusInProgress(db, &msg); err != nil {
				log.Printf("Failed to set IN-PROGRESS status for message ID %d: %v", msg.ID, err)
				return
			}

			// Proceed with processing the message
			if err := processScheduledMessage(&msg); err != nil {
				log.Printf("Error processing message ID %d: %v", msg.ID, err)
			}
		}(message)
	}
	wg.Wait()
}

func scanAndProcessConditionalMessages() {
	log.Println("Scanning & Processing conditional messages...")

	db, err := config.GetDBConnection()
	if err != nil {
		log.Println("Error getting database connection:", err)
		return
	}
	// Fetch conditional messages
	nowPlusOneSecond := time.Now().Unix() + 1
	conditionalMessages, err := messageQueueRepository.FindByStatusAndNextRetryAndRetryCountAndIsDLQ(db, string(models.PENDING), string(models.CONDITIONAL), false, models.AppConfig.DlqMessageLimit, nowPlusOneSecond)
	if err != nil {
		log.Printf("Error fetching conditional messages: %v", err)
		return
	}

	var wg sync.WaitGroup
	for _, message := range conditionalMessages {
		wg.Add(1)
		go func(msg models.MessageQueue) {
			defer wg.Done()

			// Parse the frequency cron expression
			schedule, err := cron.ParseStandard(msg.Frequency)
			if err != nil {
				log.Printf("Invalid cron expression for message ID %d: %v", msg.ID, err)
				return
			}

			// Calculate the next run time from the next retry time
			nextRetryTime := time.Unix(msg.NextRetry, 0).UTC()
			nextRun := schedule.Next(nextRetryTime)
			if time.Now().UTC().Before(nextRun) {
				return
			}

			// Acquire distributed lock
			lock := zkclient.NewDistributedLock(config.ZkConn, zkclient.LockBasePath, zkclient.LockName+strconv.Itoa(int(msg.ID)))
			acquired, err := lock.Acquire()
			if err != nil {
				log.Printf("Error acquiring lock for message ID %d: %v", msg.ID, err)
				return
			}

			if !acquired {
				// Lock not acquired, another process is already processing this message
				return
			}
			defer lock.Release()

			// Update the message status to IN_PROGRESS in the database
			if err := setMessageStatusInProgress(db, &msg); err != nil {
				log.Printf("Failed to set IN-PROGRESS status for message ID %d: %v", msg.ID, err)
				return
			}

			// Time to process the message
			log.Printf("Processing message ID %d", msg.ID)
			if err := processConditionalMessage(&msg); err != nil {
				log.Printf("Error processing conditional message ID %d: %v", msg.ID, err)
			}
		}(message)
	}
	wg.Wait()
}
