package dockercompose

// Opt is a configuration option to initialize a client
type Opt func(*Kompose) error

func WithSuppressWarnings() Opt {
	return func(k *Kompose) error {
		k.suppressWarnings = true
		return nil
	}
}

func WithVerboseOutput() Opt {
	return func(k *Kompose) error {
		k.verbose = true
		return nil
	}
}

func WithErrorOnWarning() Opt {
	return func(k *Kompose) error {
		k.errorOnWarning = true
		return nil
	}
}

type Kompose struct {
	suppressWarnings bool
	verbose          bool
	errorOnWarning   bool
}

func NewClient(opts ...Opt) (*Kompose, error) {
	k := &Kompose{
		suppressWarnings: false,
		verbose:          false,
		errorOnWarning:   false,
	}
	for _, op := range opts {
		if err := op(k); err != nil {
			return nil, err
		}
	}
	return k, nil
}
