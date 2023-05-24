package apiclient

func Retry(attempts int, f func() error) error {
	for {
		err := f()
		if err == nil {
			return nil
		}

		if attempts <= 0 {
			return err
		}

		attempts--
	}
}

func Retry3(f func() error) error {
	return Retry(3, f)
}

func Retry5(f func() error) error {
	return Retry(5, f)
}
