// Package alert provides webhook-based alerting for procwatch.
//
// Use NewSender to create a configured Sender, then call Send with a Payload
// describing the event (e.g. a process crash or resource threshold breach).
//
// Example:
//
//	s := alert.NewSender("https://hooks.example.com/procwatch")
//	err := s.Send(alert.Payload{
//		Process: "myapp",
//		Event:   "crash",
//		PID:     4321,
//		Message: "process exited with code 1",
//	})
//	if err != nil {
//		log.Printf("alert failed: %v", err)
//	}
package alert
