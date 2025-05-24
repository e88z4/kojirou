# Kojirou Kobo Support - Project Specification

## Overview

This specification outlines the extension of the Kojirou manga downloader to support Kobo e-reader formats. Currently, Kojirou downloads manga from MangaDex and creates Kindle-formatted ebooks (MOBI). This project will add support for EPUB and KEPUB formats to make the content accessible on Kobo readers.

## Goals

1. Add support for EPUB and KEPUB formats
2. Maintain the existing workflow and user experience
3. Ensure high-quality output for manga content
4. Preserve right-to-left reading direction when appropriate
5. Implement a flexible format selection system

## Format Specifications

### EPUB Format

- Standard e-book format with HTML/CSS content
- Fixed layout for manga pages
- Proper chapter structure and navigation
- Metadata including title, author, language

### KEPUB Format

- Enhanced EPUB format specific to Kobo devices
- Post-processing approach converting from EPUB
- Special Kobo-specific enhancements:
  - `kobo:` namespaced elements and attributes
  - Kobo-specific metadata in OPF file
  - CSS optimizations for Kobo readers
  - File extension change to `.kepub.epub`

## Implementation Components

### 1. EPUB Generator Module

Location: `cmd/formats/epub/epub.go`

Key features:
- Uses `go-epub` library
- Preserves manga page layout and reading direction
- Adds proper metadata and CSS
- Handles image processing similar to MOBI generator

Function signature:
```go
// GenerateEPUB creates an EPUB file from manga data
func GenerateEPUB(manga mangadex.Manga, widepage WidepagePolicy, crop bool, ltr bool) (*epub.Epub, error)
```

### 2. KEPUB Post-Processor Module

Location: `cmd/formats/epub/kepub.go`

Key features:
- Extracts and modifies EPUB content
- Adds Kobo-specific enhancements:
  - Processes HTML content to add Kobo spans
  - Updates OPF file with Kobo metadata
  - Optimizes CSS for Kobo devices
- Preserves original content and structure

Function signature:
```go
// ConvertToKEPUB transforms a standard EPUB into Kobo's KEPUB format
func ConvertToKEPUB(epubBook *epub.Epub) ([]byte, error)
```



### 4. Business Logic Updates

Location: Various files in `cmd/` directory

Key changes:
- Format selection through CLI flags
- Updated volume handling for multiple formats
- Extended directory structure for different formats
- Progress reporting for all formats

## File Organization

```
cmd/formats/
  ├── kindle/ (existing)
  │   └── mobi.go
  └── epub/
      ├── epub.go      # EPUB generation
      └── kepub.go     # KEPUB post-processing
```

## Command Line Interface

New flags:
```
--formats string    Output formats (comma-separated: mobi,epub,kepub) (default "mobi")
```

## Dependencies

- `github.com/go-epub/epub` - EPUB generation

- Existing project dependencies

## Implementation Details

### EPUB Generation

1. Create EPUB structure with proper manga metadata
2. Process images with appropriate sizing and format
3. Create HTML pages with proper layout for manga
4. Add navigation and chapter structure
5. Configure reading direction based on manga origin

### KEPUB Post-processing

1. Extract EPUB content to temporary location
2. Process HTML files to add Kobo-specific markup:
   - Add `kobo:` namespace to HTML tag
   - Add Kobo-specific image spans
   - Configure fixed layout properties
3. Update OPF file with Kobo metadata
4. Repackage with `.kepub.epub` extension



### Business Logic Changes

1. Parse format selection from command line
2. Add handlers for each format type
3. Update progress reporting for all formats
4. Extend directory management for multiple formats

## Backward Compatibility

The extension will maintain backward compatibility by:
1. Keeping MOBI as the default format
2. Preserving all existing command line options
3. Using the same directory structure pattern

## Future Extensions

1. Additional Kobo-specific optimizations
2. Support for additional image formats
3. Enhanced metadata support