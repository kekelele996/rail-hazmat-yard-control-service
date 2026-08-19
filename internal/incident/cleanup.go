package incident

func Finalize(primary error, closeFn func() error) error {
	if closeFn == nil {
		return primary
	}
	if err := closeFn(); err != nil {
		return MergeErrors(primary, err)
	}
	return primary
}
