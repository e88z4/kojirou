# Kojirou Example Commands

This document provides practical example commands for common Kojirou usage scenarios.

## Basic Usage Examples

### Generate EPUB from a manga directory

```bash
kojirou --formats=epub /path/to/manga
```

### Generate KEPUB for Kobo devices

```bash
kojirou --formats=kepub /path/to/manga
```

### Generate both EPUB and KEPUB formats

```bash
kojirou --formats=epub,kepub /path/to/manga
```

### Specify output directory

```bash
kojirou --formats=epub --destination=/path/to/output /path/to/manga
```

## Configuration Examples

### Set reading direction to left-to-right (Western style)

```bash
kojirou --formats=epub --ltr /path/to/manga
```

### Split wide pages into two separate pages

```bash
kojirou --formats=epub --widepage=split /path/to/manga
```

### Enable automatic image cropping

```bash
kojirou --formats=epub --crop /path/to/manga
```

### Combine multiple options

```bash
kojirou --formats=epub,kepub --widepage=split --crop --destination=/path/to/output /path/to/manga
```

## Batch Processing Examples

### Process all manga in a directory

```bash
for manga in /path/to/manga_collection/*; do
  kojirou --formats=epub,kepub --destination=/path/to/output "$manga"
done
```

### Process manga with different format options

```bash
# Process all manga with EPUB format
for manga in /path/to/manga_collection/*; do
  kojirou --formats=epub --destination=/path/to/epub_output "$manga"
done

# Process all manga with KEPUB format and wide page splitting
for manga in /path/to/manga_collection/*; do
  kojirou --formats=kepub --widepage=split --destination=/path/to/kepub_output "$manga"
done
```

## Advanced Examples

### High-resolution output for newer e-readers

```bash
kojirou --formats=epub --resolution=1600x2400 /path/to/manga
```

### Skip existing files (don't overwrite)

```bash
kojirou --formats=epub --noclobber /path/to/manga
```

### Set parallel processing threads

```bash
kojirou --formats=epub --parallel=4 /path/to/manga
```

### Complete example with multiple optimizations

```bash
kojirou --formats=epub,kepub \
  --widepage=split \
  --crop \
  --resolution=1280x1920 \
  --parallel=8 \
  --destination=/path/to/output \
  /path/to/manga
```

## Useful Shell Functions

Add these to your `.bashrc` or `.zshrc` for convenient processing:

```bash
# Convert single manga to EPUB
function manga2epub() {
  kojirou --formats=epub --destination="$HOME/ebooks" "$1"
}

# Convert single manga to KEPUB
function manga2kepub() {
  kojirou --formats=kepub --destination="$HOME/ebooks" "$1"
}

# Convert all manga in current directory
function convert_all_manga() {
  for dir in */; do
    kojirou --formats=epub,kepub --destination="$HOME/ebooks" "$dir"
  done
}
```

Usage:
```bash
manga2epub "/path/to/manga"
manga2kepub "/path/to/manga"
cd /path/to/manga_collection && convert_all_manga
```