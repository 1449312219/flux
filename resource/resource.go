package resource

// For the minute we just care about
type Resource interface {
	ResourceID() string // name, to correlate with what's in the cluster
	Source() string     // where did this come from (informational)
	Bytes() []byte      // the definition, for sending to platform.Sync
}
