package main

import (
	"CronServerMbs/functions"
	"CronServerMbs/scheduler"
	"github.com/jasonlvhit/gocron"
	"github.com/joho/godotenv"
	"log"
	"os"
	"strconv"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
	IntervalTimeOutboxWA, err := strconv.ParseUint(os.Getenv("INTERVAL_TIME_OUTBOX_WA"), 10, 64)
	IntervalTimeNotifWA, err := strconv.ParseUint(os.Getenv("INTERVAL_TIME_NOTIF_WA"), 10, 64)
	IntervalTimeOutboxEmail, err := strconv.ParseUint(os.Getenv("INTERVAL_TIME_OUTBOX_EMAIL"), 10, 64)
	IntervalTimeNotifEmail, err := strconv.ParseUint(os.Getenv("INTERVAL_TIME_NOTIF_EMAIL"), 10, 64)
	//IntervalBroadcastTagihan := functions.ParseTimeScheduler(os.Getenv("INTERVAL_BROADCAST_TAGIHAN"))
	//IntervalBroadcastUlangTahun := functions.ParseTimeScheduler(os.Getenv("INTERVAL_BROADCAST_ULTAH"))

	if err != nil {
		log.Fatal("Error loading .env file")
	}
	functions.Logger().Info("Starting Scheduler Cron Server MBS")
	if uint64(IntervalTimeOutboxWA) == 0 {
	}else{
		gocron.Every(uint64(IntervalTimeOutboxWA)).Seconds().Do(scheduler.CekOutboxWA)
		gocron.Every(uint64(IntervalTimeNotifWA)).Seconds().Do(scheduler.CekNotifikasiWA)
	}
	if uint64(IntervalTimeOutboxEmail) == 0 {
	}else{
		gocron.Every(uint64(IntervalTimeOutboxEmail)).Seconds().Do(scheduler.CekOutboxEmail)
		gocron.Every(uint64(IntervalTimeNotifEmail)).Seconds().Do(scheduler.CekNotifikasiEmail)
	}//gocron.Every(1).Day().At(string(IntervalBroadcastTagihan)).Do(scheduler.BroadcastTagihan)
	//gocron.Every(1).Day().At(string(IntervalBroadcastUlangTahun)).Do(scheduler.BroadcastUltah)
	<-gocron.Start()
}
