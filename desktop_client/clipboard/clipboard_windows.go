//go:build windows

package clipboard

/*
#include <windows.h>
#include <stdlib.h>
#include <string.h>

// Reads either an image (PNG preferred) or fallback text from the clipboard.
// Returns a malloc'd buffer and MIME type string. Caller must free both.
void* ReadClipboard(int* outLength, char** outMimeType) {
    if (!OpenClipboard(NULL)) {
        return NULL;
    }

    // Try PNG first (registered format)
    UINT pngFormat = RegisterClipboardFormat("PNG");
    if (IsClipboardFormatAvailable(pngFormat)) {
        HANDLE hData = GetClipboardData(pngFormat);
        if (hData) {
            SIZE_T size = GlobalSize(hData);
            void* pData = GlobalLock(hData);
            if (pData && size > 0) {
                *outLength = (int)size;
                *outMimeType = strdup("image/png");
                void* buffer = malloc(size);
                memcpy(buffer, pData, size);
                GlobalUnlock(hData);
                CloseClipboard();
                return buffer;
            }
            GlobalUnlock(hData);
        }
    }

    // Try JPEG (JFIF format)
    UINT jpegFormat = RegisterClipboardFormat("JFIF");
    if (IsClipboardFormatAvailable(jpegFormat)) {
        HANDLE hData = GetClipboardData(jpegFormat);
        if (hData) {
            SIZE_T size = GlobalSize(hData);
            void* pData = GlobalLock(hData);
            if (pData && size > 0) {
                *outLength = (int)size;
                *outMimeType = strdup("image/jpeg");
                void* buffer = malloc(size);
                memcpy(buffer, pData, size);
                GlobalUnlock(hData);
                CloseClipboard();
                return buffer;
            }
            GlobalUnlock(hData);
        }
    }

    // Try GIF format
    UINT gifFormat = RegisterClipboardFormat("GIF");
    if (IsClipboardFormatAvailable(gifFormat)) {
        HANDLE hData = GetClipboardData(gifFormat);
        if (hData) {
            SIZE_T size = GlobalSize(hData);
            void* pData = GlobalLock(hData);
            if (pData && size > 0) {
                *outLength = (int)size;
                *outMimeType = strdup("image/gif");
                void* buffer = malloc(size);
                memcpy(buffer, pData, size);
                GlobalUnlock(hData);
                CloseClipboard();
                return buffer;
            }
            GlobalUnlock(hData);
        }
    }

    // Try DIB (Device Independent Bitmap)
    if (IsClipboardFormatAvailable(CF_DIB)) {
        HANDLE hData = GetClipboardData(CF_DIB);
        if (hData) {
            SIZE_T size = GlobalSize(hData);
            void* pData = GlobalLock(hData);
            if (pData && size > 0) {
                // For simplicity, we'll indicate this as a generic image
                // In a full implementation, you'd convert DIB to PNG
                *outLength = (int)size;
                *outMimeType = strdup("image/bmp");
                void* buffer = malloc(size);
                memcpy(buffer, pData, size);
                GlobalUnlock(hData);
                CloseClipboard();
                return buffer;
            }
            GlobalUnlock(hData);
        }
    }

    // Fallback to text
    if (IsClipboardFormatAvailable(CF_UNICODETEXT)) {
        HANDLE hData = GetClipboardData(CF_UNICODETEXT);
        if (hData) {
            wchar_t* pText = (wchar_t*)GlobalLock(hData);
            if (pText) {
                // Convert Unicode to UTF-8
                int utf8Len = WideCharToMultiByte(CP_UTF8, 0, pText, -1, NULL, 0, NULL, NULL);
                if (utf8Len > 0) {
                    char* utf8Text = malloc(utf8Len);
                    WideCharToMultiByte(CP_UTF8, 0, pText, -1, utf8Text, utf8Len, NULL, NULL);
                    *outLength = utf8Len - 1; // exclude null terminator
                    *outMimeType = strdup("text/plain");
                    GlobalUnlock(hData);
                    CloseClipboard();
                    return utf8Text;
                }
                GlobalUnlock(hData);
            }
        }
    }

    CloseClipboard();
    return NULL;
}

// Writes text or PNG data to the clipboard.
// Returns 0 on success, -1 on error.
int WriteClipboard(void* data, int length, const char* mimeType) {
    if (!OpenClipboard(NULL)) {
        return -1;
    }

    EmptyClipboard();

    if (strcmp(mimeType, "text/plain") == 0) {
        // Convert UTF-8 to Unicode
        int wideLen = MultiByteToWideChar(CP_UTF8, 0, (char*)data, length, NULL, 0);
        if (wideLen <= 0) {
            CloseClipboard();
            return -1;
        }

        HGLOBAL hMem = GlobalAlloc(GMEM_MOVEABLE, (wideLen + 1) * sizeof(wchar_t));
        if (!hMem) {
            CloseClipboard();
            return -1;
        }

        wchar_t* pMem = (wchar_t*)GlobalLock(hMem);
        MultiByteToWideChar(CP_UTF8, 0, (char*)data, length, pMem, wideLen);
        pMem[wideLen] = L'\0';
        GlobalUnlock(hMem);

        if (!SetClipboardData(CF_UNICODETEXT, hMem)) {
            GlobalFree(hMem);
            CloseClipboard();
            return -1;
        }
    } else if (strcmp(mimeType, "image/png") == 0) {
        UINT pngFormat = RegisterClipboardFormat("PNG");

        HGLOBAL hMem = GlobalAlloc(GMEM_MOVEABLE, length);
        if (!hMem) {
            CloseClipboard();
            return -1;
        }

        void* pMem = GlobalLock(hMem);
        memcpy(pMem, data, length);
        GlobalUnlock(hMem);

        if (!SetClipboardData(pngFormat, hMem)) {
            GlobalFree(hMem);
            CloseClipboard();
            return -1;
        }
    } else if (strcmp(mimeType, "image/jpeg") == 0) {
        UINT jpegFormat = RegisterClipboardFormat("JFIF");

        HGLOBAL hMem = GlobalAlloc(GMEM_MOVEABLE, length);
        if (!hMem) {
            CloseClipboard();
            return -1;
        }

        void* pMem = GlobalLock(hMem);
        memcpy(pMem, data, length);
        GlobalUnlock(hMem);

        if (!SetClipboardData(jpegFormat, hMem)) {
            GlobalFree(hMem);
            CloseClipboard();
            return -1;
        }
    } else if (strcmp(mimeType, "image/gif") == 0) {
        UINT gifFormat = RegisterClipboardFormat("GIF");

        HGLOBAL hMem = GlobalAlloc(GMEM_MOVEABLE, length);
        if (!hMem) {
            CloseClipboard();
            return -1;
        }

        void* pMem = GlobalLock(hMem);
        memcpy(pMem, data, length);
        GlobalUnlock(hMem);

        if (!SetClipboardData(gifFormat, hMem)) {
            GlobalFree(hMem);
            CloseClipboard();
            return -1;
        }
    } else {
        CloseClipboard();
        return -1; // unsupported type
    }

    CloseClipboard();
    return 0;
}
*/
import "C"

import (
	"errors"
	"os/exec"
	"unsafe"
)

func readClipboard() ([]byte, string, error) {
	var length C.int
	var mimeType *C.char

	ptr := C.ReadClipboard(&length, &mimeType)
	if ptr == nil {
		return nil, "", errors.New("no supported clipboard content")
	}
	defer C.free(unsafe.Pointer(ptr))
	defer C.free(unsafe.Pointer(mimeType))

	data := C.GoBytes(unsafe.Pointer(ptr), length)
	mime := C.GoString(mimeType)

	return data, mime, nil
}

func writeClipboard(data []byte, mimeType string) error {
	cMime := C.CString(mimeType)
	defer C.free(unsafe.Pointer(cMime))

	var ptr unsafe.Pointer
	if len(data) > 0 {
		ptr = C.CBytes(data)
		defer C.free(ptr)
	}

	res := C.WriteClipboard(ptr, C.int(len(data)), cMime)
	if res != 0 {
		return errors.New("failed to write to clipboard")
	}
	return nil
}

func pasteClipboard() error {
	cmd := `Add-Type -AssemblyName System.Windows.Forms; [System.Windows.Forms.SendKeys]::SendWait('^v')`
	err := exec.Command("powershell", "-Command", cmd).Run()

	if err != nil {
		return errors.New("failed to paste clipboard")
	}

	return nil
}
