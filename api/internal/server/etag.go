package server

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/ddddami/laivan/internal/domain"
)

func setETag(w http.ResponseWriter, prefix string, id domain.ID, version int) {
	w.Header().Set("ETag", fmt.Sprintf(`"%s-%s-%d"`, prefix, id, version))
}

func parseETag(value string, prefix string, id domain.ID) (int, bool) {
	expectedPrefix := fmt.Sprintf(`"%s-%s-`, prefix, id)
	if !strings.HasPrefix(value, expectedPrefix) || !strings.HasSuffix(value, `"`) {
		return 0, false
	}

	versionText := strings.TrimSuffix(strings.TrimPrefix(value, expectedPrefix), `"`)
	version, err := strconv.Atoi(versionText)
	if err != nil || version < 1 {
		return 0, false
	}
	return version, true
}
