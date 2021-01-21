package semver

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// In order to catch parsing bugs that depend on the number or order of digits, we use ValuesGenerator
// (in values_generator.go) to create permutations of many integer values. However, as a practical
// measure to avoid very long test times, we don't use an equally large range for every value: when
// generating major, minor, and patch version numbers, we cover a 3-digit range for the major and only
// one digit for the others, on the assumption that the same numeric parsing logic is used for each.

func assertVersionComponents(t *testing.T, v Version, major int, minor int, patch int, pre string, b string) {
	assert.Equal(t, major, v.GetMajor())
	assert.Equal(t, minor, v.GetMinor())
	assert.Equal(t, patch, v.GetPatch())
	assert.Equal(t, pre, v.GetPrerelease())
	assert.Equal(t, b, v.GetBuild())
}

func parsingShouldSucceed(parseFn func(string) (Version, error), s string,
	major int, minor int, patch int, pre string, b string) func(t *testing.T) {
	return func(t *testing.T) {
		v, err := parseFn(s)
		require.NoError(t, err)
		assertVersionComponents(t, v, major, minor, patch, pre, b)
	}
}

func parsingShouldFail(parseFn func(string) (Version, error), s string) func(t *testing.T) {
	return func(t *testing.T) {
		v, err := parseFn(s)
		assert.Error(t, err)
		assert.Equal(t, Version{}, v)
	}
}

func TestParseStrict(t *testing.T) {
	strictParsingTests := func(t *testing.T, parseFn func(string) (Version, error)) {
		parsingTestsForAnyMode(t, parseFn)

		NewValuesGenerator().AddValue(0, 199).AddValue(0, 9).TestAll2(t, func(t *testing.T, a, b int) {
			t.Run("patch version cannot be omitted", parsingShouldFail(parseFn, fmt.Sprintf("%d.%d", a, b)))
		})
		NewValuesGenerator().AddValue(0, 199).TestAll1(t, func(t *testing.T, a int) {
			t.Run("micro and patch versions cannot be omitted", parsingShouldFail(parseFn, fmt.Sprintf("%d", a)))
		})
	}

	t.Run("Parse(s)", func(t *testing.T) {
		strictParsingTests(t, Parse)
	})

	t.Run("ParseAs(s, ParseModeStrict)", func(t *testing.T) {
		strictParsingTests(t, func(s string) (Version, error) { return ParseAs(s, ParseModeStrict) })
	})
}

func TestParseAllowMissingMinorAndPatch(t *testing.T) {
	parseFn := func(s string) (Version, error) { return ParseAs(s, ParseModeAllowMissingMinorAndPatch) }

	parsingTestsForAnyMode(t, parseFn)

	NewValuesGenerator().AddValue(0, 199).AddValue(0, 9).TestAll2(t, func(t *testing.T, a, b int) {
		s := fmt.Sprintf("%d.%d", a, b)

		t.Run("patch version can be omitted", parsingShouldSucceed(parseFn,
			s, a, b, 0, "", ""))

		t.Run("patch version can be omitted with prerelease", parsingShouldSucceed(parseFn,
			s+"-beta1", a, b, 0, "beta1", ""))

		t.Run("patch version can be omitted with build", parsingShouldSucceed(parseFn,
			s+"+build1", a, b, 0, "", "build1"))
	})

	NewValuesGenerator().AddValue(0, 199).TestAll1(t, func(t *testing.T, a int) {
		s := fmt.Sprintf("%d", a)

		t.Run("micro and patch versions can be omitted", parsingShouldSucceed(parseFn,
			s, a, 0, 0, "", ""))

		t.Run("micro and patch versions can be omitted with prerelease", parsingShouldSucceed(parseFn,
			s+"-beta1", a, 0, 0, "beta1", ""))

		t.Run("micro and patch versions can be omitted with build", parsingShouldSucceed(parseFn,
			s+"+build1", a, 0, 0, "", "build1"))
	})
}

func parsingTestsForAnyMode(t *testing.T, parseFn func(string) (Version, error)) {
	t.Run("valid", func(t *testing.T) {
		NewValuesGenerator().AddValue(0, 199).AddValue(0, 9).AddValue(0, 9).
			TestAll3(t, func(t *testing.T, a, b, c int) {
				s := fmt.Sprintf("%d.%d.%d", a, b, c)
				t.Run("simple complete version", parsingShouldSucceed(parseFn,
					s, a, b, c, "", ""))

				t.Run("version with prerelease (single identifier)", parsingShouldSucceed(parseFn,
					s+"-beta1", a, b, c, "beta1", ""))

				t.Run("version with build (single identifier)", parsingShouldSucceed(parseFn,
					s+"+build2", a, b, c, "", "build2"))

				t.Run("prerelease identifier of only letters", parsingShouldSucceed(parseFn,
					s+"-abc", a, b, c, "abc", ""))

				t.Run("prerelease identifier of only digits", parsingShouldSucceed(parseFn,
					s+"-123", a, b, c, "123", ""))

				t.Run("prerelease identifier of only hyphens", parsingShouldSucceed(parseFn,
					s+"----", a, b, c, "---", ""))

				t.Run("alphanumeric prerelease identifier with leading zero", parsingShouldSucceed(parseFn,
					s+"-beta1.0yes", a, b, c, "beta1.0yes", ""))

				t.Run("build identifier of only letters", parsingShouldSucceed(parseFn,
					s+"+abc", a, b, c, "", "abc"))

				t.Run("build identifier of only digits", parsingShouldSucceed(parseFn,
					s+"+123", a, b, c, "", "123"))

				t.Run("build identifier of only hyphens", parsingShouldSucceed(parseFn,
					s+"+---", a, b, c, "", "---"))

				t.Run("alphanumeric build identifier with leading zero", parsingShouldSucceed(parseFn,
					s+"+build1.0yes", a, b, c, "", "build1.0yes"))

				t.Run("leading zero in numeric build identifier", parsingShouldSucceed(parseFn,
					s+"+build1.02", a, b, c, "", "build1.02"))
			})

		NewValuesGenerator().AddValue(0, 199).AddValue(0, 9).AddValue(0, 9).AddValue(0, 2).
			TestAll4(t, func(t *testing.T, a, b, c, d int) {
				s := fmt.Sprintf("%d.%d.%d", a, b, c)
				dd := fmt.Sprintf("%d", d)
				t.Run("version with prerelease (multi-identifier)", parsingShouldSucceed(parseFn,
					s+"-beta1."+dd, a, b, c, "beta1."+dd, ""))

				t.Run("version with prerelease (multi-identifier with hyphens)", parsingShouldSucceed(parseFn,
					s+"-beta1-final."+dd, a, b, c, "beta1-final."+dd, ""))

				t.Run("version with build (multi-identifier)", parsingShouldSucceed(parseFn,
					s+"+build2."+dd, a, b, c, "", "build2."+dd))

				t.Run("version with build (multi-identifier with hyphens)", parsingShouldSucceed(parseFn,
					s+"+build2-other."+dd, a, b, c, "", "build2-other."+dd))

				t.Run("version with prerelease and build", parsingShouldSucceed(parseFn,
					s+"-beta1.rc2+build2."+dd, a, b, c, "beta1.rc2", "build2."+dd))
			})
	})

	t.Run("invalid", func(t *testing.T) {
		t.Run("non-numeric major", parsingShouldFail(parseFn, "2x.3.4"))
		t.Run("non-numeric minor", parsingShouldFail(parseFn, "2.3x.4"))
		t.Run("non-numeric patch", parsingShouldFail(parseFn, "2.3.4x"))
		t.Run("empty major", parsingShouldFail(parseFn, ".3.4"))
		t.Run("empty minor", parsingShouldFail(parseFn, "2..4"))
		t.Run("empty patch", parsingShouldFail(parseFn, "2.3."))
		t.Run("empty prerelease", parsingShouldFail(parseFn, "2.3.4-"))
		t.Run("empty prerelease identifier", parsingShouldFail(parseFn, "2.3.4-a..b"))
		t.Run("empty build", parsingShouldFail(parseFn, "2.3.4+a..b"))
		t.Run("leading zero in major", parsingShouldFail(parseFn, "02.3.4"))
		t.Run("leading zero in minor", parsingShouldFail(parseFn, "2.03.4"))
		t.Run("leading zero in patch", parsingShouldFail(parseFn, "2.3.04"))
		t.Run("leading zero in prerelease numeric identifier", parsingShouldFail(parseFn, "2.3.4-beta1.02"))
		t.Run("non-alphanumeric prerelease identifier", parsingShouldFail(parseFn, "2.3.4-beta1!"))
		t.Run("non-alphanumeric build identifier", parsingShouldFail(parseFn, "2.3.4+build!"))
		t.Run("non-ASCII character before major", parsingShouldFail(parseFn, "🔥2.3.4"))
		t.Run("non-ASCII character before minor", parsingShouldFail(parseFn, "2.🔥3.4"))
		t.Run("non-ASCII character before patch", parsingShouldFail(parseFn, "2.3.🔥4"))
		t.Run("non-ASCII character after major", parsingShouldFail(parseFn, "2🔥.3.4"))
		t.Run("non-ASCII character after minor", parsingShouldFail(parseFn, "2.3🔥.4"))
		t.Run("non-ASCII character after patch", parsingShouldFail(parseFn, "2.3.4🔥"))
		t.Run("non-ASCII character in prerelease", parsingShouldFail(parseFn, "2.3.4-beta🔥no"))
		t.Run("non-ASCII character in build", parsingShouldFail(parseFn, "2.3.4+build🔥no"))
	})
}

func TestParseAsUnknownMode(t *testing.T) {
	v, err := ParseAs("1.2.3", ParseMode(2))
	assert.Error(t, err)
	assert.Equal(t, Version{}, v)
}
