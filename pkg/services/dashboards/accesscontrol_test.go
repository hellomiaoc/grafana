package dashboards

import (
	"context"
	"fmt"
	"math/rand"
	"strconv"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	ac "github.com/grafana/grafana/pkg/services/accesscontrol"
	"github.com/grafana/grafana/pkg/services/folder"
	"github.com/grafana/grafana/pkg/services/folder/foldertest"
	"github.com/grafana/grafana/pkg/util"
)

func TestNewFolderNameScopeResolver(t *testing.T) {
	t.Run("prefix should be expected", func(t *testing.T) {
		prefix, _ := NewFolderNameScopeResolver(&FakeDashboardStore{}, NewFakeFolderStore(t), foldertest.NewFakeService())
		require.Equal(t, "folders:name:", prefix)
	})

	t.Run("resolver should convert to uid scope", func(t *testing.T) {
		dashboardStore := &FakeDashboardStore{}
		orgId := rand.Int63()
		title := "Very complex :title with: and /" + util.GenerateShortUID()

		db := &folder.Folder{Title: title, ID: rand.Int63(), UID: util.GenerateShortUID()}

		folderStore := NewFakeFolderStore(t)
		folderStore.On("GetFolderByTitle", mock.Anything, mock.Anything, mock.Anything).Return(db, nil).Once()

		scope := "folders:name:" + title

		_, resolver := NewFolderNameScopeResolver(dashboardStore, folderStore, foldertest.NewFakeService())

		resolvedScopes, err := resolver.Resolve(context.Background(), orgId, scope)
		require.NoError(t, err)
		require.Len(t, resolvedScopes, 1)

		require.Equal(t, fmt.Sprintf("folders:uid:%v", db.UID), resolvedScopes[0])

		folderStore.AssertCalled(t, "GetFolderByTitle", mock.Anything, orgId, title)
	})
	t.Run("resolver should include inherited scopes if any", func(t *testing.T) {
		dashboardStore := &FakeDashboardStore{}
		orgId := rand.Int63()
		title := "Very complex :title with: and /" + util.GenerateShortUID()

		db := &folder.Folder{Title: title, ID: rand.Int63(), UID: util.GenerateShortUID()}

		folderStore := NewFakeFolderStore(t)
		folderStore.On("GetFolderByTitle", mock.Anything, mock.Anything, mock.Anything).Return(db, nil).Once()

		scope := "folders:name:" + title

		folderSvc := foldertest.NewFakeService()
		folderSvc.ExpectedFolders = []*folder.Folder{
			{
				UID: "parent",
			},
			{
				UID: "grandparent",
			},
		}
		_, resolver := NewFolderNameScopeResolver(dashboardStore, folderStore, folderSvc)

		resolvedScopes, err := resolver.Resolve(context.Background(), orgId, scope)
		require.NoError(t, err)
		require.Len(t, resolvedScopes, 3)

		if diff := cmp.Diff([]string{
			fmt.Sprintf("folders:uid:%v", db.UID),
			"folders:uid:parent",
			"folders:uid:grandparent",
		}, resolvedScopes); diff != "" {
			t.Errorf("Result mismatch (-want +got):\n%s", diff)
		}

		folderStore.AssertCalled(t, "GetFolderByTitle", mock.Anything, orgId, title)
	})
	t.Run("resolver should fail if input scope is not expected", func(t *testing.T) {
		dashboardStore := &FakeDashboardStore{}
		_, resolver := NewFolderNameScopeResolver(dashboardStore, NewFakeFolderStore(t), foldertest.NewFakeService())

		_, err := resolver.Resolve(context.Background(), rand.Int63(), "folders:id:123")
		require.ErrorIs(t, err, ac.ErrInvalidScope)
	})
	t.Run("resolver should fail if resource of input scope is empty", func(t *testing.T) {
		dashboardStore := &FakeDashboardStore{}
		_, resolver := NewFolderNameScopeResolver(dashboardStore, NewFakeFolderStore(t), foldertest.NewFakeService())

		_, err := resolver.Resolve(context.Background(), rand.Int63(), "folders:name:")
		require.ErrorIs(t, err, ac.ErrInvalidScope)
	})
	t.Run("returns 'not found' if folder does not exist", func(t *testing.T) {
		dashboardStore := &FakeDashboardStore{}

		folderStore := NewFakeFolderStore(t)
		_, resolver := NewFolderNameScopeResolver(dashboardStore, folderStore, foldertest.NewFakeService())

		orgId := rand.Int63()
		folderStore.On("GetFolderByTitle", mock.Anything, mock.Anything, mock.Anything).Return(nil, ErrDashboardNotFound).Once()

		scope := "folders:name:" + util.GenerateShortUID()

		resolvedScopes, err := resolver.Resolve(context.Background(), orgId, scope)
		require.ErrorIs(t, err, ErrDashboardNotFound)
		require.Nil(t, resolvedScopes)
	})
}

func TestNewFolderIDScopeResolver(t *testing.T) {
	t.Run("prefix should be expected", func(t *testing.T) {
		prefix, _ := NewFolderIDScopeResolver(&FakeDashboardStore{}, NewFakeFolderStore(t), foldertest.NewFakeService())
		require.Equal(t, "folders:id:", prefix)
	})

	t.Run("resolver should convert to uid scope", func(t *testing.T) {
		dashboardStore := &FakeDashboardStore{}
		folderStore := NewFakeFolderStore(t)

		_, resolver := NewFolderIDScopeResolver(dashboardStore, folderStore, foldertest.NewFakeService())

		orgId := rand.Int63()
		uid := util.GenerateShortUID()

		db := &folder.Folder{ID: rand.Int63(), UID: uid}
		folderStore.On("GetFolderByID", mock.Anything, mock.Anything, mock.Anything).Return(db, nil).Once()

		scope := "folders:id:" + strconv.FormatInt(db.ID, 10)

		resolvedScopes, err := resolver.Resolve(context.Background(), orgId, scope)
		require.NoError(t, err)
		require.Len(t, resolvedScopes, 1)
		require.Equal(t, fmt.Sprintf("folders:uid:%v", db.UID), resolvedScopes[0])

		folderStore.AssertCalled(t, "GetFolderByID", mock.Anything, orgId, db.ID)
	})
	t.Run("resolver should should include inherited scopes if any", func(t *testing.T) {
		dashboardStore := &FakeDashboardStore{}
		folderStore := NewFakeFolderStore(t)

		folderSvc := foldertest.NewFakeService()
		folderSvc.ExpectedFolders = []*folder.Folder{
			{
				UID: "parent",
			},
			{
				UID: "grandparent",
			},
		}
		_, resolver := NewFolderIDScopeResolver(dashboardStore, folderStore, folderSvc)

		orgId := rand.Int63()
		uid := util.GenerateShortUID()

		db := &folder.Folder{ID: rand.Int63(), UID: uid}
		folderStore.On("GetFolderByID", mock.Anything, mock.Anything, mock.Anything).Return(db, nil).Once()

		scope := "folders:id:" + strconv.FormatInt(db.ID, 10)

		resolvedScopes, err := resolver.Resolve(context.Background(), orgId, scope)
		require.NoError(t, err)
		require.Len(t, resolvedScopes, 3)

		if diff := cmp.Diff([]string{
			fmt.Sprintf("folders:uid:%v", db.UID),
			"folders:uid:parent",
			"folders:uid:grandparent",
		}, resolvedScopes); diff != "" {
			t.Errorf("Result mismatch (-want +got):\n%s", diff)
		}

		folderStore.AssertCalled(t, "GetFolderByID", mock.Anything, orgId, db.ID)
	})
	t.Run("resolver should fail if input scope is not expected", func(t *testing.T) {
		dashboardStore := &FakeDashboardStore{}
		_, resolver := NewFolderIDScopeResolver(dashboardStore, NewFakeFolderStore(t), foldertest.NewFakeService())

		_, err := resolver.Resolve(context.Background(), rand.Int63(), "folders:uid:123")
		require.ErrorIs(t, err, ac.ErrInvalidScope)
	})

	t.Run("resolver should convert id 0 to general uid scope", func(t *testing.T) {
		var (
			dashboardStore = &FakeDashboardStore{}
			orgId          = rand.Int63()
			scope          = "folders:id:0"
			_, resolver    = NewFolderIDScopeResolver(dashboardStore, NewFakeFolderStore(t), foldertest.NewFakeService())
		)

		resolved, err := resolver.Resolve(context.Background(), orgId, scope)
		require.NoError(t, err)

		require.Len(t, resolved, 1)
		require.Equal(t, "folders:uid:general", resolved[0])
	})

	t.Run("resolver should fail if resource of input scope is empty", func(t *testing.T) {
		dashboardStore := &FakeDashboardStore{}
		_, resolver := NewFolderIDScopeResolver(dashboardStore, NewFakeFolderStore(t), foldertest.NewFakeService())

		_, err := resolver.Resolve(context.Background(), rand.Int63(), "folders:id:")
		require.ErrorIs(t, err, ac.ErrInvalidScope)
	})
	t.Run("returns 'not found' if folder does not exist", func(t *testing.T) {
		dashboardStore := &FakeDashboardStore{}
		folderStore := NewFakeFolderStore(t)

		_, resolver := NewFolderIDScopeResolver(dashboardStore, folderStore, foldertest.NewFakeService())

		orgId := rand.Int63()
		folderStore.On("GetFolderByID", mock.Anything, mock.Anything, mock.Anything).Return(nil, ErrDashboardNotFound).Once()

		scope := "folders:id:10"
		resolvedScopes, err := resolver.Resolve(context.Background(), orgId, scope)
		require.ErrorIs(t, err, ErrDashboardNotFound)
		require.Nil(t, resolvedScopes)
	})
}

func TestNewDashboardIDScopeResolver(t *testing.T) {
	t.Run("prefix should be expected", func(t *testing.T) {
		prefix, _ := NewDashboardIDScopeResolver(&FakeDashboardStore{}, NewFakeFolderStore(t), foldertest.NewFakeService())
		require.Equal(t, "dashboards:id:", prefix)
	})

	t.Run("resolver should convert to uid dashboard and folder scope", func(t *testing.T) {
		store := &FakeDashboardStore{}
		folderStore := NewFakeFolderStore(t)

		_, resolver := NewDashboardIDScopeResolver(store, folderStore, foldertest.NewFakeService())

		orgID := rand.Int63()
		folder := &folder.Folder{ID: 2, UID: "2"}
		dashboard := &Dashboard{ID: 1, FolderID: folder.ID, UID: "1"}

		store.On("GetDashboard", mock.Anything, mock.Anything).Return(dashboard, nil).Once()
		folderStore.On("GetFolderByID", mock.Anything, orgID, folder.ID).Return(folder, nil).Once()

		scope := ac.Scope("dashboards", "id", strconv.FormatInt(dashboard.ID, 10))
		resolvedScopes, err := resolver.Resolve(context.Background(), orgID, scope)
		require.NoError(t, err)
		require.Len(t, resolvedScopes, 2)
		require.Equal(t, fmt.Sprintf("dashboards:uid:%s", dashboard.UID), resolvedScopes[0])
		require.Equal(t, fmt.Sprintf("folders:uid:%s", folder.UID), resolvedScopes[1])
	})

	t.Run("resolver should inlude inherited scopes if any", func(t *testing.T) {
		store := &FakeDashboardStore{}
		folderStore := NewFakeFolderStore(t)

		folderSvc := foldertest.NewFakeService()
		folderSvc.ExpectedFolders = []*folder.Folder{
			{
				UID: "parent",
			},
			{
				UID: "grandparent",
			},
		}
		_, resolver := NewDashboardIDScopeResolver(store, folderStore, folderSvc)

		orgID := rand.Int63()
		folder := &folder.Folder{ID: 2, UID: "2"}
		dashboard := &Dashboard{ID: 1, FolderID: folder.ID, UID: "1"}

		store.On("GetDashboard", mock.Anything, mock.Anything).Return(dashboard, nil).Once()
		folderStore.On("GetFolderByID", mock.Anything, orgID, folder.ID).Return(folder, nil).Once()

		scope := ac.Scope("dashboards", "id", strconv.FormatInt(dashboard.ID, 10))
		resolvedScopes, err := resolver.Resolve(context.Background(), orgID, scope)
		require.NoError(t, err)
		require.Len(t, resolvedScopes, 4)

		if diff := cmp.Diff([]string{
			fmt.Sprintf("dashboards:uid:%s", dashboard.UID),
			fmt.Sprintf("folders:uid:%s", folder.UID),
			"folders:uid:parent",
			"folders:uid:grandparent",
		}, resolvedScopes); diff != "" {
			t.Errorf("Result mismatch (-want +got):\n%s", diff)
		}
	})

	t.Run("resolver should fail if input scope is not expected", func(t *testing.T) {
		_, resolver := NewDashboardIDScopeResolver(&FakeDashboardStore{}, NewFakeFolderStore(t), foldertest.NewFakeService())
		_, err := resolver.Resolve(context.Background(), rand.Int63(), "dashboards:uid:123")
		require.ErrorIs(t, err, ac.ErrInvalidScope)
	})

	t.Run("resolver should convert folderID 0 to general uid scope for the folder scope", func(t *testing.T) {
		store := &FakeDashboardStore{}
		_, resolver := NewDashboardIDScopeResolver(store, NewFakeFolderStore(t), foldertest.NewFakeService())

		dashboard := &Dashboard{ID: 1, FolderID: 0, UID: "1"}
		store.On("GetDashboard", mock.Anything, mock.Anything).Return(dashboard, nil)
		resolved, err := resolver.Resolve(context.Background(), 1, ac.Scope("dashboards", "id", "1"))
		require.NoError(t, err)

		require.Len(t, resolved, 2)
		require.Equal(t, "dashboards:uid:1", resolved[0])
		require.Equal(t, "folders:uid:general", resolved[1])
	})
}

func TestNewDashboardUIDScopeResolver(t *testing.T) {
	t.Run("prefix should be expected", func(t *testing.T) {
		prefix, _ := NewDashboardUIDScopeResolver(&FakeDashboardStore{}, NewFakeFolderStore(t), foldertest.NewFakeService())
		require.Equal(t, "dashboards:uid:", prefix)
	})

	t.Run("resolver should convert to uid dashboard and folder scope", func(t *testing.T) {
		store := &FakeDashboardStore{}
		folderStore := NewFakeFolderStore(t)
		_, resolver := NewDashboardUIDScopeResolver(store, folderStore, foldertest.NewFakeService())

		orgID := rand.Int63()
		folder := &folder.Folder{ID: 2, UID: "2"}
		dashboard := &Dashboard{ID: 1, FolderID: folder.ID, UID: "1"}

		store.On("GetDashboard", mock.Anything, mock.Anything).Return(dashboard, nil).Once()
		folderStore.On("GetFolderByID", mock.Anything, orgID, folder.ID).Return(folder, nil).Once()

		scope := ac.Scope("dashboards", "uid", dashboard.UID)
		resolvedScopes, err := resolver.Resolve(context.Background(), orgID, scope)
		require.NoError(t, err)
		require.Len(t, resolvedScopes, 2)
		require.Equal(t, fmt.Sprintf("dashboards:uid:%s", dashboard.UID), resolvedScopes[0])
		require.Equal(t, fmt.Sprintf("folders:uid:%s", folder.UID), resolvedScopes[1])
	})

	t.Run("resolver should include inherited scopes if any", func(t *testing.T) {
		store := &FakeDashboardStore{}
		folderStore := NewFakeFolderStore(t)

		folderSvc := foldertest.NewFakeService()
		folderSvc.ExpectedFolders = []*folder.Folder{
			{
				UID: "parent",
			},
			{
				UID: "grandparent",
			},
		}

		_, resolver := NewDashboardUIDScopeResolver(store, folderStore, folderSvc)

		orgID := rand.Int63()
		folder := &folder.Folder{ID: 2, UID: "2"}
		dashboard := &Dashboard{ID: 1, FolderID: folder.ID, UID: "1"}

		store.On("GetDashboard", mock.Anything, mock.Anything).Return(dashboard, nil).Once()
		folderStore.On("GetFolderByID", mock.Anything, orgID, folder.ID).Return(folder, nil).Once()

		scope := ac.Scope("dashboards", "uid", dashboard.UID)
		resolvedScopes, err := resolver.Resolve(context.Background(), orgID, scope)
		require.NoError(t, err)
		require.Len(t, resolvedScopes, 4)

		if diff := cmp.Diff([]string{
			fmt.Sprintf("dashboards:uid:%s", dashboard.UID),
			fmt.Sprintf("folders:uid:%s", folder.UID),
			"folders:uid:parent",
			"folders:uid:grandparent",
		}, resolvedScopes); diff != "" {
			t.Errorf("Result mismatch (-want +got):\n%s", diff)
		}
	})

	t.Run("resolver should fail if input scope is not expected", func(t *testing.T) {
		_, resolver := NewDashboardUIDScopeResolver(&FakeDashboardStore{}, NewFakeFolderStore(t), foldertest.NewFakeService())
		_, err := resolver.Resolve(context.Background(), rand.Int63(), "dashboards:id:123")
		require.ErrorIs(t, err, ac.ErrInvalidScope)
	})

	t.Run("resolver should convert folderID 0 to general uid scope for the folder scope", func(t *testing.T) {
		store := &FakeDashboardStore{}
		_, resolver := NewDashboardUIDScopeResolver(store, NewFakeFolderStore(t), foldertest.NewFakeService())

		dashboard := &Dashboard{ID: 1, FolderID: 0, UID: "1"}
		store.On("GetDashboard", mock.Anything, mock.Anything).Return(dashboard, nil)
		resolved, err := resolver.Resolve(context.Background(), 1, ac.Scope("dashboards", "uid", "1"))
		require.NoError(t, err)

		require.Len(t, resolved, 2)
		require.Equal(t, "dashboards:uid:1", resolved[0])
		require.Equal(t, "folders:uid:general", resolved[1])
	})
}
