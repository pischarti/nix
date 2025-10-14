package print

import (
	"fmt"
	"os"

	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/pischarti/nix/pkg/k8s"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/yaml"
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

// EventsTableWithNodes prints events with node information in a formatted table
func EventsTableWithNodes(enrichedEvents []k8s.EventWithNode) {
	t := table.NewWriter()
	t.SetOutputMirror(os.Stdout)
	t.SetStyle(table.StyleLight)

	// Check if any event has instance ID to determine if we should show that column
	hasInstanceID := false
	for _, enriched := range enrichedEvents {
		if enriched.InstanceID != "" {
			hasInstanceID = true
			break
		}
	}

	// Set table headers - include Instance ID column if any event has it
	if hasInstanceID {
		t.AppendHeader(table.Row{
			"Namespace",
			"Type",
			"Reason",
			"Object",
			"Node",
			"Instance ID",
			"Count",
			"Last Seen",
			"Message",
		})
	} else {
		t.AppendHeader(table.Row{
			"Namespace",
			"Type",
			"Reason",
			"Object",
			"Node",
			"Count",
			"Last Seen",
			"Message",
		})
	}

	// Add rows for each event
	for _, enriched := range enrichedEvents {
		event := enriched.Event
		objectRef := fmt.Sprintf("%s/%s", event.InvolvedObject.Kind, event.InvolvedObject.Name)
		lastSeen := event.LastTimestamp.Format("2006-01-02 15:04:05")

		// Truncate message if too long (shorter if we have instance ID column)
		message := event.Message
		maxLen := 70
		if hasInstanceID {
			maxLen = 60
		}
		if len(message) > maxLen {
			message = message[:maxLen-3] + "..."
		}

		// Display node name or "-" if not a pod event
		nodeName := enriched.NodeName
		if nodeName == "" {
			nodeName = "-"
		}

		if hasInstanceID {
			instanceID := enriched.InstanceID
			if instanceID == "" {
				instanceID = "-"
			}

			t.AppendRow(table.Row{
				event.Namespace,
				event.Type,
				event.Reason,
				objectRef,
				nodeName,
				instanceID,
				event.Count,
				lastSeen,
				message,
			})
		} else {
			t.AppendRow(table.Row{
				event.Namespace,
				event.Type,
				event.Reason,
				objectRef,
				nodeName,
				event.Count,
				lastSeen,
				message,
			})
		}
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

// EventsYAML prints events in YAML format
func EventsYAML(events []corev1.Event) error {
	// Convert events to YAML
	data, err := yaml.Marshal(events)
	if err != nil {
		return fmt.Errorf("failed to marshal events to YAML: %w", err)
	}

	fmt.Println(string(data))
	return nil
}
