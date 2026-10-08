package usecase

import (
	"context"

	"github.com/mondegor/go-core/errors"
	"github.com/mondegor/go-core/mrstorage"
	"github.com/mondegor/go-core/util/slices/ordered"
	"github.com/mondegor/go-core/util/xstrings"

	"print-shop-back/internal/adapter/log"
	"print-shop-back/internal/warehousing/actiongroup/back/dto"
	"print-shop-back/internal/warehousing/enum/locationkind"
	"print-shop-back/internal/warehousing/module"
)

type (
	// RefreshGroupContainers - comment struct.
	RefreshGroupContainers struct {
		storageContainer containerStorage
		storageStock     stockStorage
		logger           log.Logger
		errorWrapper     errors.Wrapper
	}

	containerStorage interface {
		FetchGroupingContainersByIDs(ctx context.Context, rowIDs []uint64) ([]dto.GroupingContainer, error)
		UpdateGroups(ctx context.Context, rows []dto.UpdateGroupContainer) error
	}

	stockStorage interface {
		FetchByLocationIDs(ctx context.Context, locationIDs []uint64, stockCursor mrstorage.IDCursor) (rows []dto.LocationStock, hasNext bool, err error)
	}
)

// NewRefreshGroupContainers - создаёт объект RefreshGroupContainers.
func NewRefreshGroupContainers(
	storageContainer containerStorage,
	storageStock stockStorage,
	logger log.Logger,
) *RefreshGroupContainers {
	return &RefreshGroupContainers{
		storageContainer: storageContainer,
		storageStock:     storageStock,
		logger:           logger,
		errorWrapper:     errors.NewServiceOperationFailedWrapper(),
	}
}

// Execute - comment method.
func (uc *RefreshGroupContainers) Execute(ctx context.Context, groupIDs []uint64) error {
	groupIDs = ordered.FilterFunc(
		groupIDs,
		func(el uint64) bool {
			return locationkind.Is(el, locationkind.Group)
		},
	)

	for _, groupID := range ordered.SortedUnique(groupIDs) {
		if err := uc.executeGroup(ctx, groupID); err != nil {
			return uc.errorWrapper.Wrap(err)
		}
	}

	return nil
}

func (uc *RefreshGroupContainers) executeGroup(ctx context.Context, groupID uint64) error {
	stocks, hasNext, err := uc.storageStock.FetchByLocationIDs(
		ctx,
		[]uint64{groupID},
		mrstorage.IDCursor{
			Limit: module.GroupContainersMax,
		},
	)
	if err != nil {
		return uc.errorWrapper.Wrap(err)
	}

	if hasNext {
		uc.logger.Error(
			ctx,
			"there are more containers in group than allowed value",
			"group_id", groupID,
			"limit", len(stocks),
		)
	}

	if len(stocks) == 0 {
		return nil
	}

	containerIDs := make([]uint64, 0, len(stocks))

	for i := range stocks {
		containerIDs = append(containerIDs, stocks[i].ContainerID)
	}

	containers, err := uc.storageContainer.FetchGroupingContainersByIDs(ctx, containerIDs)
	if err != nil {
		return uc.errorWrapper.Wrap(err)
	}

	return uc.updateContainerGroup(ctx, groupID, containers)
}

func (uc *RefreshGroupContainers) updateContainerGroup(ctx context.Context, groupID uint64, containers []dto.GroupingContainer) error {
	group := dto.UpdateGroupContainer{
		ID:   groupID,
		Tags: make([]string, 0, len(containers)),
	}

	for i, c := range containers {
		if c.Code == "" {
			return errors.ErrInternalIncorrectInputData.WithDetails("code is empty", "containerId", c.ID)
		}

		group.Tags = append(group.Tags, c.Code)

		// если у контейнера есть картинка, то выбирается только первая
		if len(c.Images) > 0 {
			if group.Images == nil {
				group.Images = make([]string, 0, len(containers)-i)
			}

			group.Images = append(group.Images, c.Images[0])
		}
	}

	group.Tags = xstrings.SortedUnique(group.Tags)
	group.Images = xstrings.SortedUnique(group.Images)

	return uc.storageContainer.UpdateGroups(ctx, []dto.UpdateGroupContainer{group})
}
