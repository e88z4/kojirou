# Kojirou Kobo Support - Implementation Checklist

This document tracks the implementation progress for adding EPUB and KEPUB format support to Kojirou.

## Phase 1: Project Setup and Basic Structure

### Step 1: Update Dependencies and Define Format Types
- [ ] Update `go.mod` to add go-epub library
- [ ] Create `cmd/formats/formats.go` file
- [ ] Define `FormatType` type with constants (MOBI, EPUB, KEPUB)
- [ ] Implement format parsing function for comma-separated format strings
- [ ] Write tests for format parsing

### Step 2: Create Directory Structure
- [ ] Create `cmd/formats/epub/` directory
- [ ] Add package documentation for each new package

### Step 3: Add Command-Line Interface
- [ ] Add `--formats` flag to root command in `cmd/root.go`
- [ ] Set default value to "mobi" for backward compatibility
- [ ] Add help text explaining supported formats (mobi,epub,kepub)
- [ ] Implement format validation
- [ ] Update PreRun hook to parse formats flag
- [ ] Write tests for format validation

## Phase 2: EPUB Generator Core

### Step 4: EPUB Generator Basic Structure
- [ ] Create `cmd/formats/epub/epub.go` with necessary imports
- [ ] Implement `GenerateEPUB` function signature
- [ ] Add metadata helper functions:
  - [ ] `mangaToTitle`
  - [ ] `mangaToLanguage`
  - [ ] `mangaToCover`
- [ ] Write tests for metadata extraction
- [ ] Write tests for basic EPUB creation

### Step 5: EPUB Page Template and Styling
- [ ] Create HTML template for manga pages
- [ ] Design CSS for proper image display
- [ ] Add support for right-to-left reading
- [ ] Implement template rendering function
- [ ] Write tests for template rendering
- [ ] Write tests for CSS application

### Step 6: EPUB Image Processing
- [ ] Implement image conversion for EPUB
- [ ] Create function to add images to EPUB
- [ ] Integrate with existing `cropAndSplit` function
- [ ] Add support for different image resolutions
- [ ] Implement proper image naming
- [ ] Write tests for image conversion
- [ ] Write tests for image addition to EPUB
- [ ] Write tests for wide page handling

### Step 7: EPUB Navigation and Structure
- [ ] Enhance `GenerateEPUB` with chapter structure
- [ ] Add table of contents generation
- [ ] Implement proper navigation points
- [ ] Add sequential page numbering
- [ ] Create nested chapter structure
- [ ] Write tests for chapter structure
- [ ] Write tests for table of contents
- [ ] Write tests for navigation functionality

## Phase 3: KEPUB Post-processor

### Step 8: KEPUB Post-processor Structure
- [x] Create `cmd/formats/epub/kepub.go` with basic structure
- [x] Implement `ConvertToKEPUB` function signature
- [x] Add utilities for EPUB extraction
- [x] Implement temporary directory management
- [x] Write tests for EPUB extraction
- [x] Write tests for directory management

### Step 9: KEPUB HTML Transformation
- [x] Implement HTML parsing functions
- [x] Add functions to find and modify text nodes
- [x] Create logic for Kobo-specific spans
- [x] Implement image element modifications
- [x] Add Kobo namespace handling
- [x] Write tests for HTML parsing
- [x] Write tests for Kobo span insertion
- [x] Write tests for namespace handling
- [x] Write tests for full document transformation

### Step 10: KEPUB Metadata Updates
- [x] Create functions to parse OPF files
- [x] Implement Kobo metadata addition
- [x] Add fixed layout property support
- [x] Set proper reading direction metadata
- [x] Ensure metadata preservation
- [x] Write tests for OPF parsing
- [x] Write tests for metadata insertion
- [x] Write tests for XML serialization
- [x] Write tests for complete OPF transformation

### Step 11: KEPUB Packaging and Finalization
- [x] Enhance `ConvertToKEPUB` to assemble final file
- [x] Implement ZIP archive creation from modified files
- [x] Add proper KEPUB extension handling
- [x] Implement temporary file cleanup
- [x] Write tests for ZIP creation
- [x] Write tests for file structure
- [x] Write tests for complete EPUB to KEPUB conversion



## Phase 5: Integration and Business Logic

### Step 15: Directory Handling for Multiple Formats
- [x] Locate `NormalizedDirectory` struct
- [x] Update `Path` method to handle different extensions
- [x] Implement `WriteEpub` method
- [x] Implement `WriteKepub` method
- [x] Update `Has` method to check for all formats
- [x] Add `HasWithExtension` method for format-specific checking
- [x] Add `GetExistingFormats` method to get all format files
- [x] Write tests for each new method
- [x] Write tests for multi-format file detection

### Step 16: Volume Processing for Multiple Formats
- [x] Locate `handleVolume` function
- [x] Modify to accept format list parameter
- [x] Implement format-specific processing paths
- [x] Add error handling for each format
- [x] Optimize for processing shared data once
- [x] Write tests for multi-format processing
- [x] Document error handling approach

### Step 17: Progress Reporting Updates
- [x] Locate progress reporting code
- [x] Add format indicators to messages
- [x] Implement parallel format reporting
- [x] Add format-specific completion messages
- [x] Update `TitledProgress` function
- [x] Write tests for format-specific messages
- [x] Write tests for multi-format progress tracking

### Step 18: Final Integration
- [x] Update main `run` function
- [x] Implement format selection handling
- [x] Connect all components
- [x] Add comprehensive error handling
- [x] Add logging for format generation
- [x] Write end-to-end tests for each format
- [x] Write tests for error cases
- [x] Write tests for edge conditions

## Testing and Quality Assurance

### Unit Tests
- [x] Ensure all new components have unit tests
- [x] Test each format generator individually
- [x] Test with various manga input data
- [x] Test error handling in all components
- [x] Fix import cycle issues and file structural problems
  - [x] Resolve package import conflicts in progress_test.go
  - [x] Fix file format issues in multi_format_implementation.go
  - [x] Fix incomplete function in kepub_integration_test.go
  - [x] Fix undefined symbols in test files
  - [x] Reorganize test files for proper compilation
- [x] Fix incomplete function implementations in test files
- [x] Fix entity/method references in test files
- [x] Create shared testhelpers package to avoid duplication
- [x] Generate and review test coverage reports

### Integration Tests
- [x] Test complete workflow with each format
- [x] Test simultaneous generation of multiple formats
- [x] Verify output file structure and content

### Manual Testing
- [ ] Test EPUB files on various e-readers
- [ ] Test KEPUB files specifically on Kobo devices
- [ ] Verify reading direction works properly
- [ ] Check that image quality is appropriate

## Documentation

### Code Documentation
- [x] Document all new public functions and types
- [x] Add package documentation for new packages
- [x] Document format-specific considerations
- [x] Create test documentation with test coverage details
- [x] Document test strategies and methodologies

### User Documentation
- [x] Update README with new format options
- [x] Document command-line flags
- [x] Add examples of generating different formats
- [x] Include troubleshooting information

## Final Steps

### Performance Optimization
- [x] Review image processing for performance bottlenecks
- [x] Optimize memory usage for large manga volumes
- [x] Consider parallel processing opportunities

### Release Preparation
- [x] Update version number
- [x] Create release notes highlighting new formats
- [x] Test installation process
- [x] Create example commands for documentation