//go:build darwin

package clipboard

/*
#cgo CFLAGS: -x objective-c
#cgo LDFLAGS: -framework Cocoa
#import <Cocoa/Cocoa.h>
#import <Foundation/Foundation.h>
#import <AppKit/AppKit.h>
#include <stdlib.h>
#include <string.h>

// Reads either an image (PNG preferred) or fallback text from the clipboard.
// Returns a malloc'd buffer and MIME type string. Caller must free both.
void* ReadClipboard(int* outLength, char** outMimeType) {
    @autoreleasepool {
        NSPasteboard *pb = [NSPasteboard generalPasteboard];
        NSArray *types = [pb types];

        // Try PNG first
        if ([types containsObject:NSPasteboardTypePNG]) {
            NSData *data = [pb dataForType:NSPasteboardTypePNG];
            if (data) {
                *outLength = (int)[data length];
                *outMimeType = strdup("image/png");
                void *buffer = malloc(*outLength);
                memcpy(buffer, [data bytes], *outLength);
                return buffer;
            }
        }

        // Try JPEG
        if ([types containsObject:@"public.jpeg"]) {
            NSData *data = [pb dataForType:@"public.jpeg"];
            if (data) {
                *outLength = (int)[data length];
                *outMimeType = strdup("image/jpeg");
                void *buffer = malloc(*outLength);
                memcpy(buffer, [data bytes], *outLength);
                return buffer;
            }
        }

        // Try GIF
        if ([types containsObject:@"com.compuserve.gif"]) {
            NSData *data = [pb dataForType:@"com.compuserve.gif"];
            if (data) {
                *outLength = (int)[data length];
                *outMimeType = strdup("image/gif");
                void *buffer = malloc(*outLength);
                memcpy(buffer, [data bytes], *outLength);
                return buffer;
            }
        }

        // Fallback to TIFF -> PNG conversion
        if ([types containsObject:NSPasteboardTypeTIFF]) {
            NSImage *img = [[NSImage alloc] initWithPasteboard:pb];
            if (img) {
                NSBitmapImageRep *rep = [[NSBitmapImageRep alloc] initWithData:[img TIFFRepresentation]];
                NSData *pngData = [rep representationUsingType:NSBitmapImageFileTypePNG properties:@{}];
                if (pngData) {
                    *outLength = (int)[pngData length];
                    *outMimeType = strdup("image/png");
                    void *buffer = malloc(*outLength);
                    memcpy(buffer, [pngData bytes], *outLength);
                    return buffer;
                }
            }
        }

        // Fallback to plain text
        NSString *text = [pb stringForType:NSPasteboardTypeString];
        if (text) {
            NSData *utf8 = [text dataUsingEncoding:NSUTF8StringEncoding];
            *outLength = (int)[utf8 length];
            *outMimeType = strdup("text/plain");
            void *buffer = malloc(*outLength);
            memcpy(buffer, [utf8 bytes], *outLength);
            return buffer;
        }

        return NULL;
    }
}
// Writes text or PNG data to the clipboard.
// Returns 0 on success, -1 on error.
int WriteClipboard(void* data, int length, const char* mimeType) {
    @autoreleasepool {
        NSPasteboard *pb = [NSPasteboard generalPasteboard];
        [pb clearContents];

        NSData *nsData = [NSData dataWithBytes:data length:length];
        NSString *type = [NSString stringWithUTF8String:mimeType];

        if ([type isEqualToString:@"text/plain"]) {
            NSString *text = [[NSString alloc] initWithData:nsData encoding:NSUTF8StringEncoding];
            if (!text) return -1;
            BOOL success = [pb setString:text forType:NSPasteboardTypeString];
            return success ? 0 : -1;
        } else if ([type isEqualToString:@"image/png"]) {
            BOOL success = [pb setData:nsData forType:NSPasteboardTypePNG];
            return success ? 0 : -1;
        } else if ([type isEqualToString:@"image/jpeg"]) {
            BOOL success = [pb setData:nsData forType:@"public.jpeg"];
            return success ? 0 : -1;
        } else if ([type isEqualToString:@"image/gif"]) {
            BOOL success = [pb setData:nsData forType:@"com.compuserve.gif"];
            return success ? 0 : -1;
        }

        return -1; // unsupported type
    }
}
*/
import "C"

import (
	"errors"
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
