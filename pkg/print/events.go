package print

import (
	"fmt"
	"os"

	"github.com/jedib0t/go-pretty/v6/table"
	corev1 "k8s.io/api/core/v1"
)

// EventsTable prints events in a formatted table
func EventsTable(events []corev1.Event) {
	t := table.NewWriter()
	t.SetOutputMirror(os.Stdout)
	t.SetStyle(table.StyleLight)

	// Set table headers
	t.AppendHeader(table.Row{
		"Namespace",
		"Type",
		"Reason",
		"Object",
		"Count",
		"Last Seen",
		"Message",
	})

	// Add rows for each event
	for _, event := range events {
		objectRef := fmt.Sprintf("%s/%s", event.InvolvedObject.Kind, event.InvolvedObject.Name)
		lastSeen := event.LastTimestamp.Format("2006-01-02 15:04:05")

		// Truncate message if too long
		message := event.Message
		if len(message) > 80 {
			message = message[:77] + "..."
		}

		t.AppendRow(table.Row{
			event.Namespace,
			event.Type,
			event.Reason,
			objectRef,
			event.Count,
			lastSeen,
			message,
		})
	}

	t.Render()
}

// EventDetailed prints a single event details in a formatted way (for detailed view)
func EventDetailed(event corev1.Event) {
	fmt.Printf("Namespace: %s\n", event.Namespace)
	fmt.Printf("Name: %s\n", event.Name)
	fmt.Printf("Type: %s\n", event.Type)
	fmt.Printf("Reason: %s\n", event.Reason)
	fmt.Printf("Object: %s/%s\n", event.InvolvedObject.Kind, event.InvolvedObject.Name)
	fmt.Printf("Count: %d\n", event.Count)
	fmt.Printf("First Seen: %s\n", event.FirstTimestamp.Format("2006-01-02 15:04:05"))
	fmt.Printf("Last Seen: %s\n", event.LastTimestamp.Format("2006-01-02 15:04:05"))
	fmt.Printf("Message: %s\n", event.Message)
	fmt.Println("---")
}
