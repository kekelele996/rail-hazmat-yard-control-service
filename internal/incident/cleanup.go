package incident

func Finalize(primary error, closeFn func() error) error {
	if closeFn == nil {
		return primary
	}
	return MergeErrors(primary, closeFn())
}
