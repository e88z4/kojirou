# Release Notes - Kojirou v1.2.0

## New Format Support: EPUB and KEPUB

We're excited to announce that Kojirou now supports generating EPUB and KEPUB formats in addition to the original MOBI format! This major update significantly expands the range of devices compatible with Kojirou-generated e-books.

### EPUB Format Support

- **Standard Compliance**: Generate EPUB 3.0 compliant files
- **Universal Compatibility**: Works with most e-readers on the market
- **Reading Direction**: Support for both left-to-right and right-to-left reading
- **Image Processing**: Enhanced image handling, including splitting of wide pages

### KEPUB Format Support

- **Kobo Optimized**: Special format with enhanced features for Kobo e-readers
- **Improved Reading Experience**: Better page turns and features on Kobo devices
- **Reading Statistics**: Support for Kobo's reading statistics and highlighting
- **Simple Conversion**: Automatically convert from EPUB to KEPUB

## Format Selection

Select which formats to generate using the new `--formats` flag:

```bash
kojirou --formats=epub,kepub path/to/manga
```

You can specify one or more formats, separated by commas. If no format is specified, MOBI is used as the default for backward compatibility.

## Performance Improvements

- **Parallel Format Generation**: Generate multiple formats simultaneously
- **Memory Optimization**: Improved memory usage for large manga volumes
- **Image Processing**: More efficient image processing pipeline
- **Disk Usage**: Optimized temporary file management

## Documentation Enhancements

- **Format Documentation**: Detailed information about format-specific features
- **Command Reference**: Complete documentation of all command-line options
- **Example Commands**: Practical examples for common usage scenarios
- **Troubleshooting Guide**: Solutions for common issues
- **Performance Optimization**: Guidelines for optimizing performance

## Other Improvements

- **Error Handling**: More robust error handling throughout the codebase
- **Progress Reporting**: Enhanced progress reporting with format-specific updates
- **Testing**: Comprehensive test suite with improved coverage
- **Code Organization**: Better code structure with shared components

## Breaking Changes

None! This update is fully backward compatible with previous versions.

## Installation

### Prebuilt Binaries

Download the latest release from the [releases page](https://github.com/leotaku/kojirou/releases).

### From Source

```bash
go install github.com/leotaku/kojirou@latest
```

## Acknowledgments

Thanks to all contributors who helped make this release possible!