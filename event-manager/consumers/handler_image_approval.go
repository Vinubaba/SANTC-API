package consumers

import (
	"context"
	"path"

	"github.com/Vinubaba/SANTC-API/common/api"
	"github.com/Vinubaba/SANTC-API/common/log"
	"github.com/Vinubaba/SANTC-API/common/storage"
	"github.com/Vinubaba/SANTC-API/event-manager/shared"

	"github.com/pkg/errors"
)

const (
	imageApprovalEventType = "imageApproval"
)

type ImageApprovalHandler struct {
	Storage   storage.Storage   `inject:""`
	Config    *shared.AppConfig `inject:""`
	Logger    *log.Logger       `inject:""`
	ApiClient api.Client        `inject:""`
}

func (h *ImageApprovalHandler) CanHandle(event Event) bool {
	return event.Type == imageApprovalEventType
}

func (h *ImageApprovalHandler) Name() string {
	return imageApprovalEventType
}

func (h *ImageApprovalHandler) Handle(ctx context.Context, event Event) error {
	if event.ImageApproval == nil {
		return errors.New("image approval is empty")
	}
	if event.ImageApproval.ChildId == "" {
		return errors.New("childId is mandatory")
	}
	if event.ImageApproval.Image == "" {
		return errors.New("image is empty")
	}

	child, err := h.ApiClient.GetChild(ctx, event.ImageApproval.ChildId)
	if err != nil {
		return errors.Wrap(err, "failed to get child")
	}

	filename, err := h.Storage.Store(ctx, event.Image, path.Join("daycares", *child.DaycareId, "children", *child.Id))
	if err != nil {
		return errors.Wrap(err, "failed to store image")
	}

	if err := h.ApiClient.AddImageApprovalRequest(ctx, api.PhotoRequestTransport{
		ChildId:     &event.ChildId,
		PublishedBy: &event.SenderId,
		Filename:    &filename,
	}); err != nil {
		return errors.Wrap(err, "failed to perform AddImageApprovalRequest")
	}

	return nil
}
