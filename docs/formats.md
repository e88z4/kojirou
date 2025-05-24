# EPUB and KEPUB Format Considerations

This document outlines important considerations when working with EPUB and KEPUB formats in Kojirou.

## EPUB Format

### Structure and Standards
- Kojirou generates EPUB 3.0 compliant files
- All EPUB files include proper navigation documents and metadata
- Files follow the standard EPUB container structure

### Image Processing
- Images are processed to maintain quality while optimizing for e-readers
- Wide pages can be handled according to user preference:
  - **Preserve**: Keep wide pages as-is (default)
  - **Split**: Split wide pages into two separate pages
  - **Scale**: Scale down wide pages to fit standard dimensions
- Autocrop option removes unnecessary borders from images

### Reading Direction
- Left-to-right (Western) and right-to-left (Japanese/manga style) reading directions are supported
- Reading direction affects:
  - Page progression direction
  - Navigation order
  - Page layout

### Limitations
- Some e-readers may have limited support for certain EPUB 3.0 features
- Image quality settings are optimized for common e-readers but may need adjustment for specific devices
- Very large manga volumes may require significant processing time

## KEPUB Format

### Kobo-Specific Enhancements
- KEPUBs include Kobo-specific markup for improved reading experience
- Text nodes are wrapped in `<span>` elements with Kobo attributes
- Additional metadata is added for Kobo device compatibility

### Conversion Process
- Standard EPUB files are converted to KEPUB format
- The conversion preserves all content and structure
- The conversion adds Kobo-specific enhancements without modifying the original content

### Device Compatibility
- KEPUB files are optimized for Kobo e-readers
- Features like page turns, highlights, and dictionary lookups work better with KEPUB files
- Some advanced features (like reading statistics) are only available with KEPUB format

### Limitations
- KEPUB files are specific to Kobo devices and may not work on other e-readers
- The conversion process may increase file size slightly
- Some complex layouts may not render exactly the same after conversion

## Recommendations for Best Results

1. Use high-quality source images for best output
2. For manga, right-to-left reading direction is usually preferred
3. Test different wide page handling options to see which works best for your content
4. Generate both EPUB and KEPUB if targeting Kobo devices
5. For very large manga series, consider processing volumes individually