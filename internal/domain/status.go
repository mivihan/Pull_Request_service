package domain

type PRStatus string

const (
	PRStatusOpen   PRStatus = "OPEN"
	PRStatusMerged PRStatus = "MERGED"
)

func (s PRStatus) IsValid() bool {
	return s == PRStatusOpen || s == PRStatusMerged
}

func (s PRStatus) String() string {
	return string(s)
}
