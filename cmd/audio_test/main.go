package main

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"fisherevans.com/project/f/internal/game/audio"
	_ "fisherevans.com/project/f/internal/resources"
	"github.com/gopxl/beep/v2/speaker"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func main() {
	// Setup logging
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
	zerolog.SetGlobalLevel(zerolog.DebugLevel)

	// Get global audio system (initialized in init())
	fmt.Println("Getting audio system...")
	sys := audio.GetSystem()
	defer speaker.Close()

	// Preload the test sound
	soundName := "song_sample"

	// Start playing the test sound
	fmt.Printf("\nPlaying '%s' at volume 0.8 (looping)...\n", soundName)
	ctrl := sys.PlaySoundOnBus(soundName, sys.Buses.Music, 0.8, &audio.PlaybackOptions{
		Loop: true,
	})

	if ctrl == nil {
		log.Fatal().Msg("failed to create sound control")
	}

	fmt.Println("\nSound Control Test CLI")
	fmt.Println("======================")
	printHelp()

	scanner := bufio.NewScanner(os.Stdin)
	for {
		fmt.Print("\n> ")
		if !scanner.Scan() {
			break
		}

		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		parts := strings.Fields(line)
		cmd := parts[0]

		switch cmd {
		case "help", "h":
			printHelp()

		case "stop":
			fmt.Println("Stopping...")
			if len(parts) >= 2 {
				ms, err := strconv.Atoi(parts[1])
				if err != nil {
					fmt.Printf("Invalid duration: %v\n", err)
					continue
				}
				duration := time.Duration(ms) * time.Millisecond
				ctrl.FadeOutAndStop(duration)
			} else {
				ctrl.Stop()
			}

		case "stop_immediate", "si":
			fmt.Println("Stopping immediately...")
			ctrl.StopImmediately()

		case "pause":
			fmt.Println("Pausing...")
			if len(parts) >= 2 {
				ms, err := strconv.Atoi(parts[1])
				if err != nil {
					fmt.Printf("Invalid duration: %v\n", err)
					continue
				}
				duration := time.Duration(ms) * time.Millisecond
				ctrl.FadeOutAndPause(duration)
			} else {
				ctrl.Pause()
			}

		case "pause_immediate", "pi":
			fmt.Println("Pausing immediately...")
			ctrl.PauseImmediately()

		case "resume":
			fmt.Println("Resuming...")
			if len(parts) >= 2 {
				ms, err := strconv.Atoi(parts[1])
				if err != nil {
					fmt.Printf("Invalid duration: %v\n", err)
					continue
				}
				duration := time.Duration(ms) * time.Millisecond
				ctrl.ResumeAndFadeIn(duration)
			} else {
				ctrl.Resume()
			}

		case "resume_immediate", "ri":
			fmt.Println("Resuming immediately...")
			ctrl.ResumeImmediately()

		case "volume", "v":
			if len(parts) < 2 {
				fmt.Println("Usage: volume <0.0-1.0>")
				continue
			}
			vol, err := strconv.ParseFloat(parts[1], 64)
			if err != nil {
				fmt.Printf("Invalid volume: %v\n", err)
				continue
			}
			fmt.Printf("Setting volume to %.2f...\n", vol)
			ctrl.SetVolume(vol)

		case "fade_out", "fo":
			duration := 1000 * time.Millisecond
			if len(parts) >= 2 {
				ms, err := strconv.Atoi(parts[1])
				if err != nil {
					fmt.Printf("Invalid duration: %v\n", err)
					continue
				}
				duration = time.Duration(ms) * time.Millisecond
			}
			fmt.Printf("Fading out over %v and pausing...\n", duration)
			ctrl.FadeOutAndPause(duration)

		case "fade_in", "fi":
			duration := 1000 * time.Millisecond
			if len(parts) >= 2 {
				ms, err := strconv.Atoi(parts[1])
				if err != nil {
					fmt.Printf("Invalid duration: %v\n", err)
					continue
				}
				duration = time.Duration(ms) * time.Millisecond
			}
			fmt.Printf("Resuming and fading in over %v...\n", duration)
			ctrl.ResumeAndFadeIn(duration)

		case "status", "s":
			fmt.Printf("Playing: %v\n", ctrl.IsPlaying())
			fmt.Printf("Paused: %v\n", ctrl.IsPaused())
			fmt.Printf("Position: %v\n", ctrl.Position())
			fmt.Printf("Duration: %v\n", ctrl.Duration())
			fmt.Printf("Remaining: %v\n", ctrl.TimeRemaining())

		case "restart":
			fmt.Println("Restarting sound...")
			ctrl.StopImmediately()
			time.Sleep(100 * time.Millisecond)
			ctrl = sys.PlaySoundOnBus(soundName, sys.Buses.Music, 0.8, &audio.PlaybackOptions{
				Loop: true,
			})
			fmt.Println("Sound restarted")

		case "quit", "q", "exit":
			fmt.Println("Exiting...")
			ctrl.StopImmediately()
			return

		default:
			fmt.Printf("Unknown command: %s\n", cmd)
			fmt.Println("Type 'help' for available commands")
		}
	}
}

func printHelp() {
	fmt.Println("\nAvailable commands:")
	fmt.Println("  help, h                    - Show this help")
	fmt.Println("  stop                       - Fade out and stop")
	fmt.Println("  stop_immediate, si         - Stop immediately")
	fmt.Println("  pause                      - Fade out and pause")
	fmt.Println("  pause_immediate, pi        - Pause immediately")
	fmt.Println("  resume                     - Fade in and resume")
	fmt.Println("  resume_immediate, ri       - Resume immediately")
	fmt.Println("  volume <0.0-1.0>, v        - Set volume")
	fmt.Println("  fade_out [ms], fo          - Fade out and pause (default 1000ms)")
	fmt.Println("  fade_in [ms], fi           - Resume and fade in (default 1000ms)")
	fmt.Println("  status, s                  - Show playback status")
	fmt.Println("  restart                    - Restart the sound")
	fmt.Println("  quit, q, exit              - Exit the program")
}
