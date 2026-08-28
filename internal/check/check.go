package check

type Kind string

const (
	KindHTTP Kind = "http"
	KindDNS  Kind = "dns"
	KindTCP  Kind = "tcp"
	KindTLS  Kind = "tls"
)

func (k Kind) Valid() bool {
	switch k {
	case KindHTTP, KindDNS, KindTCP, KindTLS:
		return true
	default:
		return false
	}
}

type Status string

const (
	StatusPending   Status = "pending"
	StatusHealthy   Status = "healthy"
	StatusWarning   Status = "warning"
	StatusUnhealthy Status = "unhealthy"
	StatusError     Status = "error"
	StatusSkipped   Status = "skipped"
)

type Outcome[T any] struct {
	Status Status
	Result T
	Err    error
}
