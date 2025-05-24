# Kojirou Command-Line Options

This document provides detailed information about all available command-line options in Kojirou.

## Format Options

### `--formats`

Specifies which output formats to generate.

```bash
kojirou --formats=epub,kepub path/to/manga
```

- **Values**: Comma-separated list of formats: `mobi`, `epub`, `kepub`
- **Default**: `mobi`
- **Example**: `--formats=mobi,epub,kepub` to generate all formats

### `--ltr`

Controls reading direction.

```bash
kojirou --ltr path/to/manga
```

- **Values**: Flag (present or absent)
- **Default**: Absent (right-to-left for manga-style reading)
- **Effect**: When present, sets left-to-right reading direction

### `--widepage`

Sets the strategy for handling wide pages.

```bash
kojirou --widepage=split path/to/manga
```

- **Values**: `preserve`, `split`, `scale`
- **Default**: `preserve`
- **Effects**:
  - `preserve`: Keep wide pages as-is
  - `split`: Split wide pages into two separate pages
  - `scale`: Scale down wide pages to fit standard dimensions

### `--crop`

Enables automatic cropping of images.

```bash
kojirou --crop path/to/manga
```

- **Values**: Flag (present or absent)
- **Default**: Absent (no cropping)
- **Effect**: When present, automatically removes borders from images

## Input/Output Options

### `--destination`

Sets the output directory.

```bash
kojirou --destination=/path/to/output path/to/manga
```

- **Value**: Path to output directory
- **Default**: Current directory
- **Example**: `--destination=/home/user/ebooks`

### `--noclobber`

Prevents overwriting existing files.

```bash
kojirou --noclobber path/to/manga
```

- **Values**: Flag (present or absent)
- **Default**: Absent (allows overwriting)
- **Effect**: When present, skips generation if output file already exists

## Processing Options

### `--parallel`

Sets the number of parallel processing threads.

```bash
kojirou --parallel=4 path/to/manga
```

- **Value**: Number of threads
- **Default**: Number of available CPU cores
- **Example**: `--parallel=8` to use 8 threads

### `--resolution`

Sets the image resolution.

```bash
kojirou --resolution=1280x1920 path/to/manga
```

- **Value**: Width x Height in pixels
- **Default**: Optimized for target format
- **Example**: `--resolution=800x1200`

## Example Commands

### Basic EPUB Generation

```bash
kojirou --formats=epub path/to/manga
```

### Generate Multiple Formats

```bash
kojirou --formats=epub,kepub --destination=/home/user/ebooks path/to/manga
```

### Manga with Custom Processing

```bash
kojirou --formats=epub --widepage=split --crop path/to/manga
```

### Western-style Comic (Left-to-Right)

```bash
kojirou --formats=epub --ltr path/to/manga
```

### High-Resolution Output

```bash
kojirou --formats=epub --resolution=1600x2400 path/to/manga
```

### Skip Existing Files

```bash
kojirou --formats=epub --noclobber path/to/manga
```

## Advanced Usage Patterns

### Processing Multiple Manga Series

```bash
for dir in /path/to/manga/*; do
  kojirou --formats=epub,kepub --destination=/path/to/output "$dir"
done
```

### Custom Processing by Format

```bash
# Generate EPUB with standard settings
kojirou --formats=epub path/to/manga

# Generate KEPUB with optimized settings for Kobo
kojirou --formats=kepub --widepage=split --crop path/to/manga
```