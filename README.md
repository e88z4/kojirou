<a href="https://mangadex.org/title/22631"><img src="./.github/header.jpg" alt="Header Image" width="80%"></a>

<h1>
  <span>Kojirou</span>
  <a href="https://goreportcard.com/report/github.com/leotaku/kojirou">
    <img src="https://goreportcard.com/badge/github.com/leotaku/kojirou?style=flat-square" alt="Go Report Card">
  </a>
  <a href="https://github.com/leotaku/kojirou/actions">
    <img src="https://img.shields.io/github/actions/workflow/status/leotaku/kojirou/check.yml?branch=master&label=check&logo=github&logoColor=white&style=flat-square" alt="Github CI Status">
  </a>
  <a href="https://github.com/leotaku/kojirou/wiki/Home">
    <img src="https://img.shields.io/github/actions/workflow/status/leotaku/kojirou/wiki.yml?branch=master&label=wiki&color=blue&logo=gitbook&logoColor=white&style=flat-square" alt="GitHub Wiki Status">
  <a/>
</h1>

> Generate perfectly formatted Kindle e-books from MangaDex manga

## Features

### Download manga and generate Kindle e-books

Kojirou will automatically download the series for the specified ID and language while outputting a folder with all the downloaded volumes.

``` shell
kojirou d86cf65b-5f6c-437d-a0af-19a31f94ec55 -l en
```

### Generate Kindle folder structure for easy synchronization

Kojirou can also output a folder structure matching that of any modern Kindle device to allow for easy synchronization using e.g. rsync.

``` shell
kojirou d86cf65b-5f6c-437d-a0af-19a31f94ec55 -l en --kindle-folder-mode
udisksctl mount -b /dev/sdb
rsync kindle/ /run/media/user/Kindle/
```

### Generate Kobo folder structure for easy synchronization

Kojirou can also output a folder structure matching that of Kobo devices for easy organization and transfer. When using the `--kobo-folder-mode` flag with KEPUB output, files are placed in `KoboBooks/<Series Title>/` and named `<Series Title> v<Volume>.kepub.epub`.

```shell
kojirou d86cf65b-5f6c-437d-a0af-19a31f94ec55 -l en --file-type=kepub --kobo-folder-mode
# or using the short flag
kojirou d86cf65b-5f6c-437d-a0af-19a31f94ec55 -l en --file-type=kepub -K
```

This will produce a structure like:

```
KoboBooks/
  My Manga Series/
    My Manga Series v01.kepub.epub
    My Manga Series v02.kepub.epub
```

You can then copy the `KoboBooks/` directory to your Kobo device.

### Customize ranking for better scantlations

Kojirou has the ability to use different [ranking algorithms](https://github.com/leotaku/kojirou/wiki/Ranking) in order to always download the highest-quality scantlations.
You can preview what would be downloaded by running in dry-run mode.

**Note:** Currently, the views and views-total ranking algorithms are broken because MangaDex no longer provides the required viewcount information.

``` shell
kojirou d86cf65b-5f6c-437d-a0af-19a31f94ec55 -l en --rank newest --dry-run
kojirou d86cf65b-5f6c-437d-a0af-19a31f94ec55 -l en --rank most
```

### Load chapters from the filesystem

Kojirou has the ability to load chapters from your local filesystem.
This can be useful if certain chapters are not available on MangaDex, or you want to convert your existing collection.
Chapters found locally are always preferred, even if they are also available on MangaDex.

``` shell
kojirou d86cf65b-5f6c-437d-a0af-19a31f94ec55 -l en --disk /path/to/directory
```

The directory structure should follow the following pattern.
Sorting of volumes, chapters and pages is done numerically and an arbitrary number of leading zeros is supported.

+ `root/`
  + `01/` :: Volume
    + `cover.{jpeg,jpg,png,bmp}` :: Volume cover (optional)
    + `01: Title/` :: Chapter (with optional title, use colon ":")
      + `01.{jpeg,jpg,png,bmp}` :: Page

### Crop whitespace from pages automatically

Kojirou has the ability to crop whitespace from the borders of manga pages.
This may be useful if your e-reader has a small screen.

``` shell
kojirou d86cf65b-5f6c-437d-a0af-19a31f94ec55 -l en --autocrop
```

### Split wide pages automatically

Kojirou has the ability to split panorama pages into two separate pages for better viewing.
It is also possible to include both the split pages and the original page.
Legal arguments to this option are "preserve", "split", "preserve-and-split" and "split-and-preserve".

``` shell
kojirou d86cf65b-5f6c-437d-a0af-19a31f94ec55 -l en --widepage=preserve-and-split
```

### Change reading direction

Kojirou, by default, generates e-books with right-to-left reading direction, as this is the default convention for most manga.
Also note that right-to-left reading does not seem to be supported on all Kindle devices.

``` shell
kojirou d86cf65b-5f6c-437d-a0af-19a31f94ec55 -l en --left-to-right
```

### Fill volume number in title

Kojirou has the ability to fill the volume number in e-book titles with an arbitrary number of leading zeros.
This is useful because Kindle devices sort titles alphabetically without any special handling of numbers.
So, for example, volume "2" would be placed before "10", while "02" would be correctly sorted.

```shell
kojirou d86cf65b-5f6c-437d-a0af-19a31f94ec55 -l en --fill-volume-number 2
```

### Use lower quality images to save space

Kojirou has the ability to download lower-quality images from MangaDex.
This can be useful to save space on your device, or to reduce the amount of data downloaded on slow or limited connections.
Legal arguments to this option are "no", "prefer" and "fallback".

```
kojirou d86cf65b-5f6c-437d-a0af-19a31f94ec55 -l en --data-saver=prefer
```

### Fallback to lower quality alternatives for broken images

MangaDex sometimes hosts images that are subtly broken and cannot be reliably converted to an image format compatible with Kindle devices.
Kojirou can be configured to fall back on reencoded lower-quality versions of these images, which often do not have the same problems.

```
kojirou d86cf65b-5f6c-437d-a0af-19a31f94ec55 -l en --data-saver=fallback
```

## Format Support

Kojirou now supports multiple output formats:

- **MOBI**: Standard format for Kindle devices (default)
- **EPUB**: Standard e-book format supported by most e-readers
- **KEPUB**: Kobo-specific format with enhanced reading features

## Format Selection

Use the `--file-type` flag to specify which formats to generate:

```bash
kojirou --file-type=mobi,epub path/to/manga
```

You can specify one or more formats, separated by commas. If no format is specified, MOBI is used as the default.

### Format-Specific Features

Each format has specific characteristics and features:

#### MOBI
- Best compatibility with Kindle devices
- Fixed image sizing optimized for Kindle screens
- Built-in support for manga-style (right-to-left) reading

#### EPUB
- Universal compatibility with most e-readers
- Support for both left-to-right and right-to-left reading
- Follows EPUB 3.0 standards
- Various image processing options

#### KEPUB
- Enhanced reading experience on Kobo devices
- Better page turn performance
- Support for Kobo's reading statistics and other features
- Based on EPUB with Kobo-specific enhancements

For more details about format-specific considerations, see [Format Documentation](docs/formats.md).

## Advanced Format Options

### Reading Direction

Control reading direction with the `--ltr` flag:

```bash
kojirou --file-type=epub --ltr path/to/manga
```

- With `--ltr`: Left-to-right reading (Western style)
- Without `--ltr`: Right-to-left reading (Manga style, default)

### Wide Page Handling

Control how wide pages are processed with the `--widepage` flag:

```bash
kojirou --file-type=epub --widepage=split path/to/manga
```

Options:
- `preserve`: Keep wide pages as-is (default)
- `split`: Split wide pages into two separate pages
- `scale`: Scale down wide pages to fit standard dimensions

### Image Processing

Control image processing with the `--crop` flag:

```bash
kojirou --file-type=epub --crop path/to/manga
```

This automatically removes unnecessary borders from images.

## Documentation

For more detailed information, refer to these documentation files:

- [Command-Line Options](docs/commands.md): Complete list of all available command-line options
- [Example Commands](docs/examples.md): Practical examples for common usage scenarios
- [Format Considerations](docs/formats.md): Details about EPUB, KEPUB, and format-specific features
- [Testing Documentation](docs/testing.md): Information about test coverage and methodology
- [Troubleshooting Guide](docs/troubleshooting.md): Solutions for common issues
- [Performance Optimization](docs/performance.md): Guidelines for optimizing performance

## Prebuilt binaries

Prebuilt binaries for Linux, Windows and MacOS on x86 and ARM processors are provided.
Visit the [release tab](https://github.com/leotaku/kojirou/releases) to download the archive for your respective setup.

On Linux and MacOS you will have to make the provided binary executable after extracting it from the archive.

``` shell
chmod u+x ./kojirou.exe
```

Afterwards, verify your installation succeeded by executing the application on the command line.

``` shell
./kojirou.exe --version
```

## Install from source

Kojirou can be installed from source easily if you already have access to a Go toolchain.
Otherwise, follow the [Go installation instructions](https://go.dev/doc/install) for your operating system, then execute the following command.

``` shell
go install github.com/leotaku/kojirou@latest
```

Afterwards, verify your installation succeeded by executing the application on the command line.

``` shell
kojirou --version
```

On many systems, the Go binary directory is not added to the list of directories searched for executables by default.
If you get a "command not found" or similar error after the previous command, run the following command and try again.
If you are using Windows, please find out how to add directories to the lookup path yourself, as there does not seem to be any quality documentation that I could link here.

``` shell
export PATH="$PATH:$(go env GOPATH)/bin"
```

## License

[MIT](./LICENSE) © Leo Gaskin 2020-2025
