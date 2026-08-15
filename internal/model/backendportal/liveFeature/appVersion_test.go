package liveFeature_test

import (
	"testing"

	"github.com/paper-indonesia/pivot-backoffice/internal/model/backendportal/liveFeature"
	"github.com/stretchr/testify/assert"
)

func TestAppVersion_Equals(t *testing.T) {
	testCases := []struct {
		desc     string
		a        liveFeature.AppVersion
		b        liveFeature.AppVersion
		expected bool
	}{
		{
			desc: "equal versions",
			a: liveFeature.AppVersion{
				Versions: map[string]string{"1.0": "stable", "2.0": "beta"},
			},
			b: liveFeature.AppVersion{
				Versions: map[string]string{"1.0": "stable", "2.0": "beta"},
			},
			expected: true,
		},
		{
			desc: "different versions",
			a: liveFeature.AppVersion{
				Versions: map[string]string{"1.0": "stable"},
			},
			b: liveFeature.AppVersion{
				Versions: map[string]string{"1.0": "stable", "2.0": "beta"},
			},
			expected: false,
		},
		{
			desc: "empty versions",
			a: liveFeature.AppVersion{
				Versions: map[string]string{},
			},
			b: liveFeature.AppVersion{
				Versions: map[string]string{},
			},
			expected: true,
		},
		{
			desc: "nil versions",
			a: liveFeature.AppVersion{
				Versions: nil,
			},
			b: liveFeature.AppVersion{
				Versions: nil,
			},
			expected: true,
		},
		{
			desc: "one nil, one empty",
			a: liveFeature.AppVersion{
				Versions: nil,
			},
			b: liveFeature.AppVersion{
				Versions: map[string]string{},
			},
			expected: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			got := tc.a.Equals(&tc.b)
			assert.Equal(t, tc.expected, got)
		})
	}
}
