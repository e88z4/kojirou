# Step-by-Step Implementation Plan for Kojirou Kobo Support

## Initial Analysis and Planning

Before diving into the implementation, let's break down the project into clear, manageable steps that build incrementally. The specification outlines two main components: EPUB generator and KEPUB post-processor, along with business logic updates to support these new formats.

### High-Level Components

1. **Command-line interface updates** - Add format selection flags
2. **EPUB generator** - Create EPUB files from manga data
3. **KEPUB post-processor** - Convert EPUBs to Kobo-specific format
4. **Business logic integration** - Update volume handling to support new formats

## Breaking Down Into Implementable Steps

Let's break each component into smaller, iterative steps:

### Phase 1: Project Setup and Command-Line Interface

1. Update `go.mod` to add necessary dependencies (e.g., go-epub)
2. Add format type definitions and command-line flags for format selection
3. Create basic package structure for new formats

### Phase 2: EPUB Generator Implementation

1. Create basic EPUB structure and metadata handling
2. Implement image processing for EPUB format
3. Add chapter and page navigation
4. Configure reading direction and layout
5. Unit test EPUB generation

### Phase 3: KEPUB Post-Processor Implementation

1. Create utility functions for extracting EPUB content
2. Implement HTML processing for Kobo-specific markup
3. Add OPF file updates for Kobo metadata
4. Implement repackaging functions
5. Unit test KEPUB conversion

### Phase 5: Business Logic Integration

1. Extend NormalizedDirectory to handle multiple formats
2. Update volume handling logic to support format selection
3. Implement progress reporting for new formats
4. Integration testing for full workflow

## Fine-Grained Implementation Steps

Now let's break these down into even smaller, testable increments:

### Phase 1: Project Setup and Command-Line Interface

#### Step 1.1: Add Dependencies and Format Type Definitions
- Update `go.mod` to add go-epub dependency
- Create format type constants and helpers

#### Step 1.2: Command-Line Flag Implementation
- Add format selection flag to root command
- Implement format parsing from comma-separated string

#### Step 1.3: Package Structure Creation
- Create directory structure for new format packages
- Add package documentation

### Phase 2: EPUB Generator Implementation

#### Step 2.1: Basic EPUB Generator Structure
- Create basic `GenerateEPUB` function signature
- Implement metadata extraction from manga data

#### Step 2.2: EPUB Page Template
- Design HTML template for manga pages
- Create CSS for proper image display

#### Step 2.3: Image Processing
- Implement image conversion and optimization
- Adapt existing cropping/splitting functions for EPUB

#### Step 2.4: Chapter Structure
- Add chapter organization
- Implement table of contents generation

#### Step 2.5: EPUB Configuration
- Configure reading direction based on manga origin
- Set fixed layout properties

#### Step 2.6: Unit Tests for EPUB Generation
- Test metadata extraction
- Test image processing
- Test chapter organization

### Phase 3: KEPUB Post-Processor Implementation

#### Step 3.1: EPUB Extraction Utilities
- Create functions to extract EPUB content
- Implement temporary directory management

#### Step 3.2: HTML Processing for Kobo
- Add functions to parse HTML documents
- Implement Kobo-specific tag insertion

#### Step 3.3: OPF File Updates
- Create OPF file parsing functions
- Add Kobo metadata insertion

#### Step 3.4: KEPUB Packaging
- Implement repackaging function
- Add proper extension handling

#### Step 3.5: Unit Tests for KEPUB Conversion
- Test HTML processing
- Test OPF modifications
- Test full EPUB to KEPUB conversion

### Phase 4: CBZ Generator Implementation

#### Step 4.1: Basic CBZ Structure
- Create `GenerateCBZ` function signature
- Implement ZIP archive creation

#### Step 4.2: Image Handling
- Add functions to convert and optimize images for CBZ
- Implement proper naming convention for sequential reading

#### Step 4.3: Chapter Organization
- Add directory structure for chapters
- Implement metadata for comic readers

#### Step 4.4: Unit Tests for CBZ Generation
- Test ZIP structure
- Test image processing
- Test chapter organization

### Phase 5: Business Logic Integration

#### Step 5.1: Directory Handling Updates
- Extend NormalizedDirectory for new formats
- Add file extension handling

#### Step 5.2: Volume Processing Updates
- Modify volume handling to support format selection
- Add format-specific error handling

#### Step 5.3: Progress Reporting
- Update progress tracking for new formats
- Add format-specific messages

#### Step 5.4: Integration Tests
- Test full workflow with multiple formats
- Test error handling and edge cases

## Final Implementation Plan

After further review, here's the optimized implementation plan broken down into small, testable units:

### Phase 1: Project Setup and Basic Structure

#### Step 1: Update Dependencies and Define Format Types
- Update `go.mod` with new dependencies
- Create format type constants in a new file

#### Step 2: Create Directory Structure
- Create new package directories
- Add package documentation files

#### Step 3: Add Command-Line Interface
- Add format selection flag to root command
- Implement format parsing logic
- Write tests for format parsing

### Phase 2: EPUB Generator Core

#### Step 4: EPUB Generator Basic Structure
- Create EPUB generator interface
- Implement basic metadata extraction
- Write tests for metadata handling

#### Step 5: EPUB Page Template and Styling
- Create HTML template for pages
- Design CSS for proper manga display
- Write tests for template rendering

#### Step 6: EPUB Image Processing
- Implement image conversion for EPUB
- Adapt existing image processing functions
- Write tests for image handling

#### Step 7: EPUB Navigation and Structure
- Add chapter and navigation structure
- Implement table of contents
- Write tests for navigation structure

### Phase 3: KEPUB Post-processor

#### Step 8: KEPUB Post-processor Structure
- Create basic KEPUB conversion function
- Implement EPUB extraction utilities
- Write tests for extraction process

#### Step 9: KEPUB HTML Transformation
- Add HTML parsing and modification
- Implement Kobo-specific tag insertion
- Write tests for HTML transformations

#### Step 10: KEPUB Metadata Updates
- Create OPF file handling functions
- Add Kobo metadata insertion
- Write tests for metadata modifications

#### Step 11: KEPUB Packaging and Finalization
- Implement repackaging functionality
- Add proper extension handling
- Write tests for full conversion process



### Phase 5: Integration and Business Logic

#### Step 15: Directory Handling for Multiple Formats
- Extend NormalizedDirectory for new formats
- Add file extension handling
- Write tests for directory operations

#### Step 16: Volume Processing for Multiple Formats
- Update volume handling logic
- Add format-specific processing paths
- Write tests for volume processing

#### Step 17: Progress Reporting Updates
- Extend progress tracking for new formats
- Add format-specific messages
- Write tests for progress reporting

#### Step 18: Final Integration
- Connect all components in the business logic
- Implement format selection from command line
- Write end-to-end tests for full workflow

## LLM Prompts for Implementing Each Step

### Prompt 1: Project Setup and Format Types

```
I'm implementing Kobo e-reader format support for a manga downloader called Kojirou. The first step is to set up the project structure and define format types.

1. Update the go.mod file to add the go-epub library (github.com/go-epub/epub) as a dependency
2. Create a new file at cmd/formats/formats.go to define format types
3. Define a FormatType type and constants for MOBI, EPUB, KEPUB, and CBZ formats
4. Create a function to parse format types from a comma-separated string
5. Add tests for the format parsing function

The existing project structure has:
- cmd/formats/ - Package for format-specific code
- cmd/formats/kindle/ - Kindle format generation
- mangadex/ - MangaDex API client and manga data structures

Please implement this first step with appropriate tests.
```

### Prompt 2: Command-Line Interface Update

```
Now I need to update the command-line interface to support format selection. The current application uses the Cobra library for CLI commands.

1. Add a new global flag "--formats" to the root command in cmd/root.go
2. Default value should be "mobi" for backward compatibility
3. Add help text explaining the supported formats (mobi,epub,kepub,cbz)
4. Update the command's PreRun hook to parse the formats flag
5. Store the parsed formats for later use
6. Add validation to ensure only supported formats are specified
7. Add tests for the format validation

The existing root command structure is in cmd/root.go and uses the spf13/cobra library.

Please implement this change with appropriate tests.
```

### Prompt 3: Create Package Structure

```
Let's create the basic package structure for our new format generators.

1. Create a new directory cmd/formats/epub/
2. Create a new directory cmd/formats/cbz/
3. Create package declarations in both with appropriate documentation
4. Create basic file structures:
   - cmd/formats/epub/epub.go
   - cmd/formats/epub/kepub.go
   - cmd/formats/cbz/cbz.go
5. Define basic interfaces and type signatures
6. Add appropriate imports

For each file, create the package declaration, add godoc comments explaining the purpose, and define the main function signatures from the specification:

For epub.go:
- func GenerateEPUB(manga mangadex.Manga, widepage WidepagePolicy, crop bool, ltr bool) (*epub.Epub, error)

For kepub.go:
- func ConvertToKEPUB(epubBook *epub.Epub) ([]byte, error)

For cbz.go:
- func GenerateCBZ(manga mangadex.Manga, widepage WidepagePolicy, crop bool, ltr bool) (*bytes.Buffer, error)

The WidepagePolicy type should be imported from the kindle package.

Please implement these basic package structures with appropriate documentation.
```

### Prompt 4: EPUB Generator Basic Structure

```
Now let's implement the basic structure of the EPUB generator. We'll start with the core functionality to handle manga metadata and create an EPUB structure.

1. In cmd/formats/epub/epub.go, implement the following:
   - Import necessary packages (go-epub, image handling, etc.)
   - Define constants for page templates and CSS
   - Implement the GenerateEPUB function
   - Create helper functions for metadata extraction

2. The GenerateEPUB function should:
   - Create a new EPUB instance
   - Set metadata (title, author, language)
   - Add basic CSS for page layout
   - Return the EPUB object

3. Helper functions should include:
   - mangaToTitle - Extract title from manga data
   - mangaToLanguage - Extract language from manga data
   - mangaToCover - Extract cover image from manga data

4. Add tests for:
   - Metadata extraction
   - Basic EPUB structure creation

Base this implementation on the patterns used in cmd/formats/kindle/mobi.go, adapting them for the EPUB format.

Please implement the EPUB generator basic structure with tests.
```

### Prompt 5: EPUB Page Template and Styling

```
Let's implement the page template and styling for the EPUB format. This will define how manga pages are displayed in the EPUB.

1. In cmd/formats/epub/epub.go, enhance the existing constants:
   - Create a detailed HTML template for manga pages
   - Define CSS for proper image display in e-readers
   - Add support for right-to-left reading

2. The HTML template should:
   - Use semantic HTML5
   - Support fixed layout for manga pages
   - Include proper image embedding

3. The CSS should:
   - Ensure images scale properly on different screen sizes
   - Support both left-to-right and right-to-left reading
   - Optimize for e-reader display

4. Add tests for:
   - Template rendering
   - CSS application

Use the existing pageTemplateString in cmd/formats/kindle/mobi.go as reference, but adapt it for EPUB standards.

Please implement the EPUB page templates and styling with tests.
```

### Prompt 6: EPUB Image Processing

```
Now let's implement image processing for EPUB format. We need to handle manga images properly for optimal display on e-readers.

1. In cmd/formats/epub/epub.go, implement:
   - Functions to process images for EPUB inclusion
   - Reuse existing image processing from kindle package
   - Add EPUB-specific image optimization

2. Specifically implement:
   - A function to convert images to appropriate format (JPEG)
   - A function to add images to the EPUB
   - Integration with existing cropAndSplit function (from kindle package)

3. The image processing should:
   - Ensure proper resolution for e-readers
   - Maintain aspect ratios
   - Support image splitting for wide pages
   - Support both left-to-right and right-to-left ordering

4. Add tests for:
   - Image conversion
   - Image addition to EPUB
   - Wide page handling

Reference the existing image processing in cmd/formats/kindle/mobi.go but adapt it for EPUB requirements.

Please implement the EPUB image processing functionality with tests.
```

### Prompt 7: EPUB Navigation and Structure

```
Let's implement the navigation and chapter structure for the EPUB format. This will organize the manga into chapters and pages.

1. In cmd/formats/epub/epub.go, enhance the GenerateEPUB function to:
   - Create a proper chapter structure
   - Generate a table of contents
   - Add navigation points

2. Implement functions to:
   - Process chapters from manga data
   - Create section breaks between chapters
   - Add sequential page numbers

3. The navigation structure should:
   - Support nested chapters (volume -> chapter -> pages)
   - Include chapter titles and numbers
   - Support skipping to specific chapters

4. Add tests for:
   - Chapter structure creation
   - Table of contents generation
   - Navigation functionality

Use the chapter handling logic in cmd/formats/kindle/mobi.go as reference, but adapt it for EPUB standards.

Please implement the EPUB navigation and structure with tests.
```

### Prompt 8: KEPUB Post-processor Structure

```
Now let's create the basic structure for the KEPUB post-processor. This will convert standard EPUBs to Kobo's enhanced format.

1. In cmd/formats/epub/kepub.go, implement:
   - Basic ConvertToKEPUB function structure
   - Utilities for EPUB extraction and processing
   - Temporary directory management

2. The ConvertToKEPUB function should:
   - Take an EPUB object
   - Return a byte array of the converted KEPUB
   - Handle errors properly

3. Implement utilities for:
   - Extracting an EPUB to temporary files
   - Managing temporary directory lifecycle
   - Basic file handling

4. Add tests for:
   - EPUB extraction process
   - Temporary directory management

This step focuses on the infrastructure for KEPUB conversion without implementing the actual transformations yet.

Please implement the basic KEPUB post-processor structure with tests.
```

### Prompt 9: KEPUB HTML Transformation

```
Let's implement the HTML transformation logic for KEPUB conversion. This will add Kobo-specific markup to make the content optimized for Kobo devices.

1. In cmd/formats/epub/kepub.go, implement:
   - Functions to parse HTML documents
   - Logic to add Kobo-specific spans to text
   - Code to modify image elements
   - Functions to add Kobo namespace

2. Specifically add:
   - HTML parsing using a suitable Go HTML parser
   - Functions to find and modify text nodes
   - Image handling enhancements
   - Functions to save modified HTML

3. The transformation should:
   - Add kobo:* namespace declarations
   - Add <kobo:div> tags around key content
   - Add spans around sentences and paragraphs
   - Preserve existing attributes and structure

4. Add tests for:
   - HTML parsing
   - Kobo span insertion
   - Namespace handling
   - Full document transformation

Please implement the KEPUB HTML transformation logic with tests.
```

### Prompt 10: KEPUB Metadata Updates

```
Let's implement the metadata updates for KEPUB conversion. This involves modifying the OPF file to add Kobo-specific metadata.

1. In cmd/formats/epub/kepub.go, implement:
   - Functions to parse and modify content.opf
   - Logic to add Kobo-specific metadata elements
   - Support for fixed-layout properties

2. Specifically add:
   - XML parsing for OPF files
   - Functions to add Kobo metadata items
   - Preservation of existing metadata
   - Support for manga-specific properties

3. The metadata updates should include:
   - Adding kobo:* namespace
   - Setting proper page progression direction
   - Adding fixed layout properties
   - Setting appropriate reading direction

4. Add tests for:
   - OPF parsing
   - Metadata insertion
   - XML serialization
   - Full OPF transformation

Please implement the KEPUB metadata update functionality with tests.
```

### Prompt 11: KEPUB Packaging and Finalization

```
Let's implement the final packaging and finalization of the KEPUB format. This will assemble all the components into the final .kepub.epub file.

1. In cmd/formats/epub/kepub.go, enhance ConvertToKEPUB to:
   - Assemble modified files into a ZIP archive
   - Use the correct .kepub.epub extension
   - Handle cleanup of temporary files
   - Return the final byte array

2. Implement functions for:
   - Creating ZIP archives from directory structures
   - Proper file ordering and compression
   - Finalizing the KEPUB file

3. The packaging should:
   - Maintain proper file structure
   - Use appropriate compression settings
   - Include all modified files
   - Generate valid KEPUB that Kobo devices can read

4. Add tests for:
   - ZIP creation
   - File structure validation
   - Complete EPUB to KEPUB conversion

Please implement the KEPUB packaging functionality with tests.
```



### Prompt 15: Directory Handling for Multiple Formats

```
Let's update the directory handling to support multiple formats. This involves extending the NormalizedDirectory struct in the kindle package.

1. Identify the NormalizedDirectory struct (likely in cmd/formats/kindle/directory.go)
2. Extend it to support multiple formats by:
   - Updating the Path method to handle different extensions
   - Adding methods for each new format (WriteEpub, WriteKepub, WriteCbz)
   - Supporting format-specific file patterns

3. Implement new methods:
   - WriteEpub for writing EPUB files
   - WriteKepub for writing KEPUB files
   - WriteCbz for writing CBZ files

4. Update the Has method to check for files of any supported format
5. Add tests for all new methods

Please implement the directory handling updates for multiple formats with tests.
```

### Prompt 16: Volume Processing for Multiple Formats

```
Let's update the volume processing logic to support multiple formats. This involves modifying the handleVolume function in the business logic.

1. Find the handleVolume function (likely in cmd/business.go)
2. Modify it to accept a list of formats to generate
3. Implement format-specific processing paths for each format
4. Add error handling for each format

5. Update the function to:
   - Process common data once (pages, metadata)
   - Generate each requested format
   - Report progress for each format
   - Handle errors for each format separately

6. Add tests for:
   - Multi-format volume processing
   - Format-specific error handling

Please implement the volume processing updates for multiple formats with tests.
```

### Prompt 17: Progress Reporting Updates

```
Let's update the progress reporting system to support multiple formats. This will provide user feedback during format generation.

1. Find the progress reporting code (likely in cmd/formats/progress.go)
2. Extend it to support format-specific messages by:
   - Adding format indicators to progress messages
   - Supporting parallel format generation reporting
   - Adding format-specific completion messages

3. Update:
   - TitledProgress function to include format information
   - Add format-specific prefixes
   - Support format-specific completion states

4. Add tests for:
   - Format-specific progress messages
   - Multi-format progress tracking

Please implement the progress reporting updates for multiple formats with tests.
```

### Prompt 18: Final Integration

```
Let's complete the implementation by integrating all components into the main business logic. This will connect the format selection to the actual generation of different formats.

1. Update the main run function in cmd/business.go to:
   - Parse format selection from command-line flags
   - Pass selected formats to handleVolume function
   - Add logging for format generation

2. Ensure all components are properly connected:
   - Format selection from CLI
   - Directory handling for all formats
   - Format generators for EPUB, KEPUB, and CBZ
   - Progress reporting for all formats

3. Add validation and error handling:
   - Validate format selection before processing
   - Handle errors from each format generator
   - Provide useful error messages

4. Add end-to-end tests:
   - Test complete workflow with multiple formats
   - Test error cases and edge conditions

Please implement the final integration connecting all components with appropriate tests.
```