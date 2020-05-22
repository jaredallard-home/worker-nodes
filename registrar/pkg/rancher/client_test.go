package rancher

import (
	"context"
	"os"
	"testing"

	"github.com/davecgh/go-spew/spew"
)

func TestGetClusterRegistrationToken(t *testing.T) {
	c := NewClient(os.Getenv("RANCHER_TOKEN"))

	d, err := c.GetClusterRegistrationToken(context.TODO(), "")
	if err != nil {
		t.Error(err)
		return
	}

	if len(d) == 0 {
		spew.Dump(d)
		t.Errorf("expected to get > 0 clusterregistrationtokens, but got %d", len(d))
		return
	}

	if d[0].Token == "" {
		t.Errorf("expected to get a token, but didn't")
	}

	d, err = c.GetClusterRegistrationToken(context.TODO(), "c-l7jc8")
	if err != nil {
		t.Error(err)
		return
	}

	if len(d) == 0 {
		spew.Dump(d)
		t.Errorf("expected to get > 0 clusterregistrationtokens, but got %d (filter)", len(d))
		return
	}

	if d[0].Token == "" {
		t.Errorf("expected to get a token, but didn't (filter)")
		return
	}

	if d[0].ClusterID != "c-l7jc8" {
		t.Errorf("expected to get filtered clusterId, but go one that isn't matching the filter: %s", d[0].ClusterID)
		return
	}
}
