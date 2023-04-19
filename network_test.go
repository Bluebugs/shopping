//go:build network
// +build network

package main

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"golang.org/x/sync/errgroup"
)

func Test_UploadDownloadYaml(t *testing.T) {
	slUploaded := &shoppingList{
		Name: "test",
		Items: []item{
			{
				What:    "unchecked",
				Checked: false,
			},
			{
				What:    "checked",
				Checked: true,
			},
		},
	}
	slDownloaded := &shoppingList{}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	code, status, err := slUploaded.uploadYaml(ctx)
	assert.Nil(t, err)
	assert.NotNil(t, status)
	assert.NotEmpty(t, code)

	var eg errgroup.Group
	eg.Go(func() error {
		return slDownloaded.downloadYaml(ctx, code)
	})

	s := <-status
	assert.True(t, s.OK)
	assert.Nil(t, s.Error)

	err = eg.Wait()
	assert.Nil(t, err)

	assert.Equal(t, slUploaded.Name, slDownloaded.Name)
	assert.Len(t, slDownloaded.Items, 2)
	assert.Equal(t, slUploaded.Items[0].What, slDownloaded.Items[0].What)
	assert.Equal(t, slUploaded.Items[0].Checked, slDownloaded.Items[0].Checked)
	assert.Equal(t, slUploaded.Items[1].What, slDownloaded.Items[1].What)
	assert.Equal(t, slUploaded.Items[1].Checked, slDownloaded.Items[1].Checked)
}
