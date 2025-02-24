package cronjobs

import (
	"github.com/robfig/cron/v3"
	"log"
)

// 2025/02/24 08:47:00 Function 3 running every 5 minutes, 2 minutes apart
// 2025/02/24 08:50:00 Function 1 running every 5 minutes
// 2025/02/24 08:51:00 Function 2 running every 5 minutes, 1 minute apart
// 2025/02/24 08:52:00 Function 3 running every 5 minutes, 2 minutes apart

func InitCronJobs() {
	log.Println("Starting Cron Jobs")
	c := cron.New()

	// Function 1: Run every 5 minutes at 0 minutes
	_, err := c.AddFunc("*/5 * * * *", func() {
		log.Println("Function 1 running every 5 minutes")
	})
	if err != nil {
		log.Println("Error scheduling Function 1:", err)
	}

	// Function 2: Run every 5 minutes at 1 minute mark
	_, err = c.AddFunc("1-59/5 * * * *", func() {
		log.Println("Function 2 running every 5 minutes, 1 minute apart")
	})
	if err != nil {
		log.Println("Error scheduling Function 2:", err)
	}

	// Function 3: Run every 5 minutes at 2 minutes mark
	_, err = c.AddFunc("2-59/5 * * * *", func() {
		log.Println("Function 3 running every 5 minutes, 2 minutes apart")
	})
	if err != nil {
		log.Println("Error scheduling Function 3:", err)
	}

	c.Start()
}
