package main

import (
	"encoding/json"
	"fmt"
	"github.com/Bishwas-py/notify"
	"github.com/godbus/dbus/v5"
	"log"
	"math/rand"
	"os"
	"os/exec"
	"path"
	"runtime"
	"strings"
	"time"
)

const (
	AppID          = "Love Thy Eyes"
	ConfigFileName = "lovethyeyes.json"
)

// UserStats tracks user engagement with the app
type UserStats struct {
	EyeLoveScore   int       `json:"eye_love_score"`
	EyeHatredScore int       `json:"eye_hatred_score"`
	BreaksTaken    int       `json:"breaks_taken"`
	BreaksSkipped  int       `json:"breaks_skipped"`
	TotalUsageTime time.Time `json:"total_usage_time"`
	StartTime      time.Time `json:"start_time"`
	LastStatShow   time.Time `json:"last_stat_show"`
}

// Notification messages for variety
type MessageBank struct {
	ShortBreakMessages []string
	LongBreakMessages  []string
	EcoMessages        []string
	StatMessages       []string
}

// EyeSaver main application struct
type EyeSaver struct {
	ShortBreakInterval time.Duration
	LongBreakInterval  time.Duration
	StatShowInterval   time.Duration
	AppIcon            string
	MessageBank        MessageBank
	Stats              UserStats
	StartTime          time.Time
	ConfigPath         string
	LastShortBreak     time.Time
	LastLongBreak      time.Time
	NotificationsOn    bool
	AudioOn            bool
}

// NewEyeSaver creates a new instance with default settings
func NewEyeSaver() *EyeSaver {
	_, filename, _, _ := runtime.Caller(0)
	appIcon := path.Join(path.Dir(filename), "logo.png")
	configDir, _ := os.UserConfigDir()
	configPath := path.Join(configDir, ConfigFileName)

	// Seed the random number generator
	rand.Seed(time.Now().UnixNano())

	saver := &EyeSaver{
		ShortBreakInterval: 7 * time.Second,  // 20-20-20 rule
		LongBreakInterval:  12 * time.Second, // Longer break every hour
		StatShowInterval:   14 * time.Minute, // Show stats every 14 minutes
		AppIcon:            appIcon,
		ConfigPath:         configPath,
		NotificationsOn:    true,
		AudioOn:            true,
		Stats: UserStats{
			StartTime:    time.Now(),
			LastStatShow: time.Now(),
		},
		MessageBank: MessageBank{
			ShortBreakMessages: []string{
				"Look 20 feet away for 20 seconds. Your eyes will thank you!",
				"Time for the 20-20-20 rule! Look away at something distant.",
				"Give your eyes a micro-vacation. Look at the horizon for a moment.",
				"Roll your eyes in circles 5 times each direction. It helps reduce strain.",
				"Blink rapidly for 15 seconds - it refreshes your tear film!",
				"Cup your palms over your closed eyes for 30 seconds. Feel the darkness?",
				"Focus on your breath for 30 seconds while looking away from the screen.",
				"Trace an imaginary figure eight with your eyes. It exercises eye muscles!",
			},
			LongBreakMessages: []string{
				"Go out, see something beautiful, something far away, a mountain, a river, a forest or a sea.",
				"Time to stretch your legs AND your eyes! A 5-minute walk outside works wonders.",
				"Get some fresh air! Look at the clouds, trees, or just enjoy the open space.",
				"Your eyes and brain need a proper break. Step outside and look at distant objects.",
				"Try the 10-10-10 exercise: look at something 10 feet, 100 feet, and 1000 feet away.",
				"Grab a cup of tea or water, and gaze out a window while you enjoy it.",
				"Find the most distant thing you can see from your window and focus on it for 2 minutes.",
				"Your eyes deserve a panoramic view - find one and take it all in for a few minutes.",
			},
			EcoMessages: []string{
				"Water your plants while giving your eyes a break!",
				"Check on your indoor garden while resting your eyes.",
				"Open a window for fresh air - good for you and saves energy on climate control!",
				"Use this break to sort some recyclables - your eyes and the planet will thank you.",
				"Turn off unnecessary lights while you take your break - energy saving is eye saving!",
				"Check that your electronic devices are on power-saving mode during your break.",
				"Go admire a tree or plant - biophilia is good for mental health and eye strain!",
				"If you have a balcony or garden, spend your break time there with nature.",
			},
			StatMessages: []string{
				"Your Eye Love-Hatred ratio is currently %d:%d. How do you feel about that?",
				"Eye Care Stats: %d breaks embraced, %d breaks ignored. Your eyes remember!",
				"Screen Time Check: You've been using your computer for %s today. Remember to hydrate!",
				"Eye Love: %d | Eye Hatred: %d | Time since last break: %s",
				"Eye Care Level: %s. Based on your %d/%d love-hate ratio.",
				"You've taken %d breaks and skipped %d. Each break matters to your eye health!",
				"Your eyes have been working for %s today. They've earned %d points of care.",
				"Screen relationship status: %s (based on your %d:%d care ratio)",
			},
		},
	}

	// Try to load saved stats
	saver.LoadStats()
	return saver
}

// SaveStats persists user statistics to disk
func (e *EyeSaver) SaveStats() error {
	e.Stats.TotalUsageTime = time.Now()
	data, err := json.MarshalIndent(e.Stats, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(e.ConfigPath, data, 0644)
}

// LoadStats retrieves user statistics from disk
func (e *EyeSaver) LoadStats() error {
	data, err := os.ReadFile(e.ConfigPath)
	if err != nil {
		if os.IsNotExist(err) {
			// First run, just use defaults
			return nil
		}
		return err
	}

	return json.Unmarshal(data, &e.Stats)
}

// Start begins the eye saver timers
func (e *EyeSaver) Start() {
	log.Println("Love Thy Eyes started!")
	log.Printf("Short breaks every %s, long breaks every %s",
		e.ShortBreakInterval, e.LongBreakInterval)

	// Set initial times
	e.LastShortBreak = time.Now()
	e.LastLongBreak = time.Now()

	// Main event loop
	ticker := time.NewTicker(1 * time.Second)
	for {
		select {
		case <-ticker.C:
			e.checkTimers()
		}
	}
}

// checkTimers evaluates if it's time for any notifications
func (e *EyeSaver) checkTimers() {
	now := time.Now()

	// Check for short break
	if now.Sub(e.LastShortBreak) >= e.ShortBreakInterval {
		e.TriggerShortBreak()
		e.LastShortBreak = now
	}

	// Check for long break
	if now.Sub(e.LastLongBreak) >= e.LongBreakInterval {
		e.TriggerLongBreak()
		e.LastLongBreak = now
	}

	// Check if it's time to show stats
	if now.Sub(e.Stats.LastStatShow) >= e.StatShowInterval {
		e.ShowStats()
		e.Stats.LastStatShow = now
	}

	// Save stats periodically
	e.SaveStats()
}

// randomMessage gets a random message from a slice
func randomMessage(messages []string) string {
	return messages[rand.Intn(len(messages))]
}

// TriggerShortBreak notifies the user to take a short eye break
func (e *EyeSaver) TriggerShortBreak() {
	log.Println("Hey")
	// Combine a regular eye message with an eco message occasionally
	useEcoMessage := rand.Intn(5) == 0 // 20% chance

	message := randomMessage(e.MessageBank.ShortBreakMessages)
	if useEcoMessage {
		message += "\n\n" + randomMessage(e.MessageBank.EcoMessages)
	}

	notification := notify.Notification{
		Title:   "Time for a quick eye break!",
		AppIcon: e.AppIcon,
		AppID:   AppID,
		Body:    message,
		Actions: notify.Actions{
			{
				Title:   "I did it! 👍",
				Trigger: func() { e.RecordBreakTaken(true) },
			},
			{
				Title:   "Snooze (5 min)",
				Trigger: func() { e.SnoozeBreak(5 * time.Minute) },
			},
		},
		Timeout: int(30 * time.Second),
	}

	// Set notification sound
	notification.SetSoundByName(notify.DialogInformation)

	//// Also use beeep for audio if enabled
	//if e.AudioOn {
	//	//beeep.Notify("Love Thy Eyes", message, e.AppIcon)
	//	notification
	//}

	log.Println("Short break triggered:", message)

	// Track notification
	notificationID, err := notification.Trigger()
	if err != nil {
		log.Printf("Error triggering notification: %v", err)
	} else {
		log.Printf("Notification ID: %d triggered", notificationID)
	}
}

// TriggerLongBreak notifies the user to take a longer eye break
func (e *EyeSaver) TriggerLongBreak() {
	message := randomMessage(e.MessageBank.LongBreakMessages)

	notification := notify.Notification{
		Title:   "Time for a longer break!",
		AppIcon: e.AppIcon,
		AppID:   AppID,
		Body:    message,
		Actions: notify.Actions{
			{
				Title:   "Break taken! 🏆",
				Trigger: func() { e.RecordBreakTaken(false) },
			},
			{
				Title:   "Snooze (10 min)",
				Trigger: func() { e.SnoozeBreak(10 * time.Minute) },
			},
			{
				Title: "Logout Now!",
				Trigger: func() {
					e.Stats.EyeLoveScore += 5 // Extra points for logout
					e.SaveStats()
					HandleLogout()
				},
			},
		},
		Timeout: int(60 * time.Second),
	}

	// Set a more urgent sound for long breaks
	notification.SetSoundByName(notify.DialogWarning)

	//// Also use beeep for audio if enabled
	//if e.AudioOn {
	//	beeep.Alert("Love Thy Eyes - Long Break", message, e.AppIcon)
	//}

	log.Printf("Long break triggered: %s", message)
	_, err := notification.Trigger()
	if err != nil {
		log.Printf("Error triggering notification: %v", err)
	}
}

// RecordBreakTaken updates statistics when a user takes a break
func (e *EyeSaver) RecordBreakTaken(isShort bool) {
	log.Println("Break taken!")
	e.Stats.BreaksTaken++
	e.Stats.EyeLoveScore++

	// Extra point for long breaks
	if !isShort {
		e.Stats.EyeLoveScore++
	}

	e.SaveStats()
}

// RecordBreakSkipped updates statistics when a user skips a break
func (e *EyeSaver) RecordBreakSkipped() {
	log.Println("Break skipped!")
	e.Stats.BreaksSkipped++
	e.Stats.EyeHatredScore++
	e.SaveStats()
}

// SnoozeBreak delays the next break
func (e *EyeSaver) SnoozeBreak(duration time.Duration) {
	log.Printf("Break snoozed for %s", duration)
	e.LastShortBreak = time.Now().Add(duration)
	e.SaveStats()
}

// ShowStats displays the current eye care statistics
func (e *EyeSaver) ShowStats() {
	log.Println("Showing stats")

	// Calculate various stats
	usageTime := time.Since(e.Stats.StartTime)
	usageHours := usageTime.Hours()
	timeSinceLastBreak := time.Since(e.LastShortBreak)

	// Determine eye care level
	var eyeCareLevel string
	ratio := 0.0
	if e.Stats.EyeHatredScore > 0 {
		ratio = float64(e.Stats.EyeLoveScore) / float64(e.Stats.EyeHatredScore)
	} else {
		ratio = float64(e.Stats.EyeLoveScore)
	}

	switch {
	case ratio >= 3:
		eyeCareLevel = "Eye Care Champion"
	case ratio >= 2:
		eyeCareLevel = "Eye Friendly User"
	case ratio >= 1:
		eyeCareLevel = "Eye Neutral User"
	case ratio > 0:
		eyeCareLevel = "Eye Strain Risk"
	default:
		eyeCareLevel = "Eye Care Beginner"
	}

	// Get relationship status
	var relationshipStatus string
	switch {
	case ratio >= 4:
		relationshipStatus = "In love with healthy eyes"
	case ratio >= 2:
		relationshipStatus = "In a healthy relationship with your screen"
	case ratio >= 1:
		relationshipStatus = "It's complicated with your screen"
	default:
		relationshipStatus = "Screen addiction alert"
	}

	// Format message from template
	statTemplate := randomMessage(e.MessageBank.StatMessages)
	var statMessage string

	switch {
	case strings.Contains(statTemplate, "Love-Hatred ratio"):
		statMessage = fmt.Sprintf(statTemplate, e.Stats.EyeLoveScore, e.Stats.EyeHatredScore)
	case strings.Contains(statTemplate, "breaks embraced"):
		statMessage = fmt.Sprintf(statTemplate, e.Stats.BreaksTaken, e.Stats.BreaksSkipped)
	case strings.Contains(statTemplate, "Screen Time Check"):
		statMessage = fmt.Sprintf(statTemplate, formatDuration(usageTime))
	case strings.Contains(statTemplate, "Time since last break"):
		statMessage = fmt.Sprintf(statTemplate, e.Stats.EyeLoveScore, e.Stats.EyeHatredScore, formatDuration(timeSinceLastBreak))
	case strings.Contains(statTemplate, "Eye Care Level"):
		statMessage = fmt.Sprintf(statTemplate, eyeCareLevel, e.Stats.EyeLoveScore, e.Stats.EyeHatredScore)
	case strings.Contains(statTemplate, "You've taken"):
		statMessage = fmt.Sprintf(statTemplate, e.Stats.BreaksTaken, e.Stats.BreaksSkipped)
	case strings.Contains(statTemplate, "Your eyes have been working"):
		statMessage = fmt.Sprintf(statTemplate, formatDuration(usageTime), e.Stats.EyeLoveScore)
	case strings.Contains(statTemplate, "Screen relationship status"):
		statMessage = fmt.Sprintf(statTemplate, relationshipStatus, e.Stats.EyeLoveScore, e.Stats.EyeHatredScore)
	default:
		statMessage = fmt.Sprintf("Eye Love: %d | Eye Hatred: %d | Breaks: %d",
			e.Stats.EyeLoveScore, e.Stats.EyeHatredScore, e.Stats.BreaksTaken)
	}

	// Show a tip based on score ratio
	var tip string
	if ratio < 1 {
		tip = "\n\nTip: Taking regular breaks improves productivity and prevents eye strain!"
	} else if usageHours > 4 && e.Stats.BreaksTaken < 5 {
		tip = "\n\nTip: For every 20 minutes, look at something 20 feet away for 20 seconds."
	}

	statMessage += tip

	notification := notify.Notification{
		Title:   "Eye Care Stats",
		AppIcon: e.AppIcon,
		AppID:   AppID,
		Body:    statMessage,
		Actions: notify.Actions{
			{
				Title: "Thanks for reminding me!",
				Trigger: func() {
					e.Stats.EyeLoveScore++
					e.SaveStats()
				},
			},
		},
		Timeout: int(15 * time.Second),
	}

	notification.SetSoundByName(notify.DialogInformation)
	_, _ = notification.Trigger()
}

// formatDuration returns a human-readable duration
func formatDuration(d time.Duration) string {
	h := int(d.Hours())
	m := int(d.Minutes()) % 60

	if h > 0 {
		return fmt.Sprintf("%dh %dm", h, m)
	}
	return fmt.Sprintf("%dm", m)
}

// ToggleNotifications enables/disables visual notifications
func (e *EyeSaver) ToggleNotifications() {
	e.NotificationsOn = !e.NotificationsOn
	log.Printf("Notifications turned %s", onOffString(e.NotificationsOn))
}

// ToggleAudio enables/disables audio alerts
func (e *EyeSaver) ToggleAudio() {
	e.AudioOn = !e.AudioOn
	log.Printf("Audio alerts turned %s", onOffString(e.AudioOn))
}

// onOffString converts boolean to on/off string
func onOffString(val bool) string {
	if val {
		return "ON"
	}
	return "OFF"
}

// HandleLogout is called when the user chooses to logout
func HandleLogout() {
	log.Println("Logout action triggered!")

	// Try GNOME logout first
	err := LogoutViaGnome()
	if err != nil {
		log.Printf("GNOME logout failed: %v\n", err)

		// Fall back to systemd
		err = LogoutViaSystemd()
		if err != nil {
			log.Printf("Systemd logout failed: %v\n", err)
		}
	}
}

// LogoutViaGnome attempts to logout via GNOME session manager
func LogoutViaGnome() error {
	conn, err := dbus.SessionBus()
	if err != nil {
		return fmt.Errorf("failed to connect to session bus: %v", err)
	}

	obj := conn.Object("org.gnome.SessionManager", "/org/gnome/SessionManager")
	call := obj.Call("org.gnome.SessionManager.Logout", 0, uint32(0))

	return call.Err
}

// LogoutViaSystemd attempts to logout via systemd loginctl
func LogoutViaSystemd() error {
	cmd := exec.Command("loginctl", "terminate-user", os.Getenv("USER"))
	return cmd.Run()
}

func main() {
	eyeSaver := NewEyeSaver()

	// You could add command line flags here for customization
	// For example:
	// flag.DurationVar(&eyeSaver.ShortBreakInterval, "short", 20*time.Minute, "Short break interval")
	// flag.Parse()

	eyeSaver.Start()
}
