# Kojirou Troubleshooting Guide

This document helps diagnose and resolve common issues when using Kojirou.

## Common Issues and Solutions

### Installation Problems

#### Issue: Error installing Kojirou
```
go: github.com/leotaku/kojirou@latest: unknown revision latest
```

**Solution:**
```bash
# Use a specific version instead
go get github.com/leotaku/kojirou@v1.0.0
```

#### Issue: Missing dependencies
```
package image/xyz: cannot find package
```

**Solution:**
```bash
# Update dependencies
go mod tidy
```

### Input/Output Errors

#### Issue: Cannot access manga directory
```
error: cannot access /path/to/manga: no such file or directory
```

**Solution:**
- Verify the path is correct
- Ensure you have read permissions for the directory
- Use absolute paths instead of relative paths

#### Issue: Cannot write to output directory
```
error: cannot create file /path/to/output/manga.epub: permission denied
```

**Solution:**
- Verify the output directory exists
- Ensure you have write permissions for the directory
- Create the output directory if it doesn't exist:
  ```bash
  mkdir -p /path/to/output
  ```

### Format Generation Issues

#### Issue: EPUB generation failed
```
error: failed to generate EPUB: invalid manga data
```

**Solution:**
- Check that the manga directory has a valid structure
- Ensure image files are in a supported format (JPG, PNG)
- Verify the manga has at least one volume with chapters and pages

#### Issue: KEPUB conversion failed
```
error: failed to convert to KEPUB: EPUB object is nil
```

**Solution:**
- Make sure EPUB generation succeeded first
- Try generating only EPUB format to verify it works
- Check for corrupt image files in your manga

### Image Processing Problems

#### Issue: Wide page handling not working correctly
```
warning: could not process wide page: aspect ratio calculation failed
```

**Solution:**
- Try a different wide page policy (`--widepage=preserve` or `--widepage=scale`)
- Check if the image files are valid and can be opened
- Consider preprocessing the images manually if they have unusual dimensions

#### Issue: Automatic cropping removing too much content
```
warning: autocrop may have removed image content
```

**Solution:**
- Disable autocropping (`--crop` flag omitted)
- Preprocess images manually if necessary
- Use higher resolution source images if available

### Performance Issues

#### Issue: Processing is very slow
```
Processing manga... (this may take a while)
```

**Solution:**
- Increase parallel processing threads (`--parallel=8`)
- Process smaller batches of manga
- Use SSD storage for input/output if possible
- Close other resource-intensive applications

#### Issue: Out of memory during processing
```
fatal error: runtime: out of memory
```

**Solution:**
- Process one volume at a time instead of full series
- Reduce parallel processing threads (`--parallel=2`)
- Reduce the resolution (`--resolution=800x1200`)
- Increase system swap space if possible

## Advanced Troubleshooting

### Checking Image Files

If you suspect image files are causing issues:

```bash
# List all image files and their sizes
find /path/to/manga -type f \( -name "*.jpg" -o -name "*.png" \) -exec ls -la {} \;

# Check for corrupt image files
find /path/to/manga -type f \( -name "*.jpg" -o -name "*.png" \) -exec identify {} \; 2>/dev/null || echo "Corrupt image found"
```

### Verbose Output

For detailed debugging information:

```bash
# Add -v flag for verbose output
kojirou -v --formats=epub /path/to/manga
```

### Checking Generated Files

If the generated EPUB/KEPUB doesn't work on your device:

```bash
# Validate EPUB structure
epubcheck /path/to/output/manga.epub

# Extract EPUB to inspect contents
mkdir extracted_epub
unzip /path/to/output/manga.epub -d extracted_epub
```

## Getting Help

If you continue to experience issues:

1. Check the [GitHub Issues](https://github.com/leotaku/kojirou/issues) for similar problems
2. Create a new issue with:
   - Kojirou version
   - Command you're running
   - Error messages
   - System information (OS, Go version)
   - Sample of the problematic manga structure (if possible)