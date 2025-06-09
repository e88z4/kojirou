package kindle

import (
	"path"
)

type KoboDirectory struct {
	baseDir string
	series  string
}

func NewKoboDirectory(baseDir, series string) KoboDirectory {
	return KoboDirectory{
		baseDir: baseDir,
		series:  series,
	}
}

func (k KoboDirectory) Path(volume string) string {
	return path.Join(k.baseDir, k.series, volume)
}
