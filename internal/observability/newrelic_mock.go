package observability

import (
	"encoding/json"
	"log"
	"time"
)

// RecordCustomEvent records a custom event (mock implementation)
func RecordCustomEvent(eventName string, attrs map[string]interface{}) {
	attrs["timestamp"] = time.Now().UTC().Format(time.RFC3339)
	attrs["eventName"] = eventName

	// In production, this would send to New Relic
	// For now, just log it
	data, _ := json.MarshalIndent(attrs, "", "  ")
	log.Printf("[OBSERVABILITY] %s\n%s", eventName, string(data))
}
