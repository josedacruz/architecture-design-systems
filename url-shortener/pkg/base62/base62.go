package base62

import "github.com/josedacruz/architecture-design-system/url-shortener/internal/config"

// toBase62 converts a decimal integer (int64) into its Base62 string representation.
// This algorithm is similar to converting to binary or hexadecimal, using division and remainders.
func ToBase62(n int64) string {
	if n == 0 {
		return string(config.Base62Chars[0]) // Special case for 0
	}

	res := make([]byte, 0) // Slice to build the result characters
	for n > 0 {
		remainder := n % int64(len(config.Base62Chars))  // Get the remainder when divided by 62
		res = append(res, config.Base62Chars[remainder]) // Append the corresponding Base62 character
		n /= int64(len(config.Base62Chars))              // Update n for the next iteration
	}

	// The characters are generated in reverse order (least significant digit first),
	// so we need to reverse the slice to get the correct Base62 string.
	for i, j := 0, len(res)-1; i < j; i, j = i+1, j-1 {
		res[i], res[j] = res[j], res[i]
	}
	return string(res) // Convert the byte slice to a string
}
