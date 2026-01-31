#!/bin/bash
# Installation script for document redaction tools

echo "================================================"
echo "Installing Document Redaction Tools"
echo "================================================"

# Update package list
sudo apt-get update

echo ""
echo "1. Installing Tesseract OCR (with German + English)"
echo "================================================"
sudo apt-get install -y tesseract-ocr tesseract-ocr-eng tesseract-ocr-deu

# Verify Tesseract
if command -v tesseract &> /dev/null; then
    echo "✅ Tesseract installed: $(tesseract --version | head -1)"
else
    echo "❌ Tesseract installation failed"
    exit 1
fi

echo ""
echo "2. Installing ImageMagick (Image manipulation)"
echo "================================================"
sudo apt-get install -y imagemagick

# Verify ImageMagick
if command -v convert &> /dev/null; then
    echo "✅ ImageMagick installed: $(convert --version | head -1)"
else
    echo "❌ ImageMagick installation failed"
    exit 1
fi

echo ""
echo "3. Installing Ghostscript (PDF processing)"
echo "================================================"
sudo apt-get install -y ghostscript

# Verify Ghostscript
if command -v gs &> /dev/null; then
    echo "✅ Ghostscript installed: $(gs --version)"
else
    echo "❌ Ghostscript installation failed"
    exit 1
fi

echo ""
echo "4. Installing Poppler Utils (PDF to image conversion)"
echo "================================================"
sudo apt-get install -y poppler-utils

# Verify pdftoppm
if command -v pdftoppm &> /dev/null; then
    echo "✅ Poppler utils installed"
else
    echo "❌ Poppler utils installation failed"
    exit 1
fi

echo ""
echo "5. Installing PDFtk (PDF toolkit - optional)"
echo "================================================"
sudo apt-get install -y pdftk || echo "⚠️  PDFtk not available, using alternatives"

echo ""
echo "================================================"
echo "✅ All required tools installed successfully!"
echo "================================================"
echo ""
echo "Installed tools:"
echo "  - Tesseract OCR: $(which tesseract)"
echo "  - ImageMagick:   $(which convert)"
echo "  - Ghostscript:   $(which gs)"
echo "  - Poppler utils: $(which pdftoppm)"
echo ""
echo "You can now use the document redaction features."
echo ""
