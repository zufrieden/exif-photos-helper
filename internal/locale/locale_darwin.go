// Package locale reports the user's preferred UI language.
package locale

/*
#cgo CFLAGS: -x objective-c
#cgo LDFLAGS: -framework Foundation
#import <Foundation/Foundation.h>
#include <stdlib.h>

// Honours both the system language list and the per-app language set in
// System Settings › General › Language & Region › Applications.
static char *preferredLanguage(void) {
	@autoreleasepool {
		NSString *lang = [[NSLocale preferredLanguages] firstObject];
		return strdup(lang ? [lang UTF8String] : "");
	}
}
*/
import "C"

import "unsafe"

// Preferred returns the first preferred language, e.g. "fr-CH" or "en-US".
func Preferred() string {
	cs := C.preferredLanguage()
	defer C.free(unsafe.Pointer(cs))
	return C.GoString(cs)
}
