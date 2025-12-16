package berio

type BerInput interface {
	Read() (string, error)
	Select(prompt string, options []string) string
}
