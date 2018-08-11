package scanner

type Mode int8

const (
	ModeFiles Mode = iota
	ModeDirectories
	ModeAll
)

type Penetration int8

const (
	PenetrationFlat Penetration = iota
	PenetrationRecursive
)

type Builder struct {
	mode        Mode
	penetration Penetration
	directories []string
	filter      Filter
}

func (b *Builder) Files() *Builder {
	b.mode = ModeFiles
	return b
}

func (b *Builder) Directories() *Builder {
	b.mode = ModeDirectories
	return b
}

func (b *Builder) Flat() *Builder {
	b.penetration = PenetrationFlat
	return b
}

func (b *Builder) Recursive() *Builder {
	b.penetration = PenetrationRecursive
	return b
}

func (b *Builder) In(directories ...string) *Builder {
	b.directories = append(b.directories, directories...)
	return b
}

func (b *Builder) Match(filter Filter) *Builder {
	b.filter = filter
	return b
}

func (b *Builder) Build() (Scanner, error) {
	var (
		scanner Scanner
		err     error
	)

	scanner, err = b.buildConcreteScanner()
	if err != nil {
		return nil, err
	}

	scanner = b.buildFilterScanner(scanner)

	return scanner, nil
}

func (b *Builder) MustBuild() Scanner {
	s, err := b.Build()
	if err != nil {
		panic(err)
	}

	return s
}

func (b *Builder) buildConcreteScanner() (Scanner, error) {
	if b.penetration == PenetrationFlat {
		switch len(b.directories) {
		case 0:
			return NewBasicScanner()
		case 1:
			return NewBasicScanner(WithDir(b.directories[0]))
		default:
			var scanners []Scanner
			for _, d := range b.directories {
				scanner, err := NewBasicScanner(WithDir(d))
				if err != nil {
					return nil, err
				}

				scanners = append(scanners, scanner)
			}
			return NewMultiScanner(scanners...), nil
		}
	}

	switch len(b.directories) {
	case 0:
		return NewRecursiveScanner()
	case 1:
		return NewRecursiveScanner(WithDirectories(b.directories[0]))
	default:
		return NewRecursiveScanner(WithDirectories(b.directories...))
	}
}

func (b *Builder) buildFilterScanner(scanner Scanner) Scanner {
	var filters []Filter

	if b.mode == ModeFiles {
		filters = append(filters, RegularFilesFilter)
	}

	if b.mode == ModeDirectories {
		filters = append(filters, DirectoriesFilter)
	}

	if b.filter != nil {
		filters = append(filters, b.filter)
	}

	switch len(filters) {
	case 0:
		return scanner
	case 1:
		return NewFilterScanner(scanner, filters[0])
	default:
		return NewFilterScanner(scanner, AndFilter(filters...))
	}
}

func NewBuilder() *Builder {
	return &Builder{
		mode:        ModeAll,
		penetration: PenetrationFlat,
	}
}
