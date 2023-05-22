//go:build network
// +build network

package main

import (
	"context"
	"testing"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/test"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
	require.Nil(t, err)
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

func Test_DownloadButtonInteraction(t *testing.T) {
	a, done := setupAppDataWithTemporaryDb()
	defer done()
	assert.NotNil(t, a)

	a.createUI()
	assert.NotNil(t, a.tabs)

	test.AssertRendersToImage(t, "create_ui.png", a.win.Canvas())
	test.AssertRendersToMarkup(t, "create_ui.markup", a.win.Canvas())

	slUploaded := &shoppingList{
		Name: "test",
		Items: []item{
			{
				What:    "uploaded-unchecked",
				Checked: false,
			},
			{
				What:    "uploaded-checked",
				Checked: true,
			},
		},
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	code, status, err := slUploaded.uploadYaml(ctx)
	require.Nil(t, err)
	assert.NotNil(t, status)
	assert.NotEmpty(t, code)

	downloadCtx, downloadCancel := context.WithCancel(context.Background())
	d, entry := a.shoppingLists[0].downloadYamlDialog(downloadCtx, downloadCancel, a.win)
	assert.NotNil(t, d)
	assert.NotNil(t, entry)

	entry.Text = code
	d.Show()

	// download button: 448,339
	test.TapCanvas(a.win.Canvas(), fyne.NewPos(448, 339))

	s := <-status
	assert.True(t, s.OK)
	assert.Nil(t, s.Error)

	downloadDone := downloadCtx.Done()
	assert.NotNil(t, downloadDone)
	<-downloadDone

	assert.Len(t, a.shoppingLists, 1)
	assert.Len(t, a.shoppingLists[0].Items, 2)
	assert.Equal(t, "uploaded-unchecked", a.shoppingLists[0].Items[0].What)
	assert.False(t, a.shoppingLists[0].Items[0].Checked)
	assert.Equal(t, "uploaded-checked", a.shoppingLists[0].Items[1].What)
	assert.True(t, a.shoppingLists[0].Items[1].Checked)
}

func Test_UploadButtonInteraction(t *testing.T) {
	a, done := setupAppDataWithTemporaryDb()
	defer done()
	assert.NotNil(t, a)

	a.createUI()
	assert.NotNil(t, a.tabs)

	test.AssertRendersToImage(t, "create_ui.png", a.win.Canvas())
	test.AssertRendersToMarkup(t, "create_ui.markup", a.win.Canvas())

	uploadCtx, uploadCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer uploadCancel()

	code := a.shoppingLists[0].uploadYamlDialog(uploadCtx, uploadCancel, a.win)
	assert.NotEmpty(t, code)

	slDownloaded := &shoppingList{}

	downloadCtx, downloadCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer downloadCancel()

	err := slDownloaded.downloadYaml(downloadCtx, code)
	assert.Nil(t, err)

	<-uploadCtx.Done()
	<-downloadCtx.Done()

	assert.Len(t, a.shoppingLists, 1)
	assert.Len(t, a.shoppingLists[0].Items, 2)
	assert.Equal(t, "unchecked", a.shoppingLists[0].Items[0].What)
	assert.False(t, a.shoppingLists[0].Items[0].Checked)
	assert.Equal(t, "checked", a.shoppingLists[0].Items[1].What)
	assert.True(t, a.shoppingLists[0].Items[1].Checked)
}
