/* SPDX-License-Identifier: MIT
 *
 * Copyright (C) 2019 WireGuard LLC. All Rights Reserved.
 */

package ui

import (
	"github.com/lxn/walk"
	"golang.zx2c4.com/wireguard/windows/service"
)

type sizeAndState struct {
	size  walk.Size
	state service.TunnelState
}

type TunnelStatusImageProvider struct {
	imagesBySizeAndState map[sizeAndState]*walk.Bitmap
	stoppedBrush         *walk.SolidColorBrush
	startingBrush        *walk.SolidColorBrush
	startedBrush         *walk.SolidColorBrush
	stoppedPen           *walk.CosmeticPen
	startingPen          *walk.CosmeticPen
	startedPen           *walk.CosmeticPen
}

func NewTunnelStatusImageProvider() (*TunnelStatusImageProvider, error) {
	tsip := &TunnelStatusImageProvider{imagesBySizeAndState: make(map[sizeAndState]*walk.Bitmap)}
	var err error

	var disposables walk.Disposables
	defer disposables.Treat()

	if tsip.stoppedBrush, err = walk.NewSolidColorBrush(walk.RGB(225, 225, 225)); err != nil {
		return nil, err
	}
	disposables.Add(tsip.stoppedBrush)

	if tsip.startingBrush, err = walk.NewSolidColorBrush(walk.RGB(254, 192, 49)); err != nil {
		return nil, err
	}
	disposables.Add(tsip.startingBrush)

	if tsip.startedBrush, err = walk.NewSolidColorBrush(walk.RGB(54, 206, 66)); err != nil {
		return nil, err
	}
	disposables.Add(tsip.startedBrush)

	if tsip.stoppedPen, err = walk.NewCosmeticPen(walk.PenSolid, walk.RGB(225-10, 225-10, 225-10)); err != nil {
		return nil, err
	}
	disposables.Add(tsip.stoppedPen)

	if tsip.startingPen, err = walk.NewCosmeticPen(walk.PenSolid, walk.RGB(254-10, 192-10, 49-10)); err != nil {
		return nil, err
	}
	disposables.Add(tsip.startingPen)

	if tsip.startedPen, err = walk.NewCosmeticPen(walk.PenSolid, walk.RGB(54-10, 206-10, 66-10)); err != nil {
		return nil, err
	}
	disposables.Add(tsip.startedPen)

	disposables.Spare()

	return tsip, nil
}

func (tsip *TunnelStatusImageProvider) Dispose() {
	if tsip.imagesBySizeAndState != nil {
		for _, img := range tsip.imagesBySizeAndState {
			img.Dispose()
		}
		tsip.imagesBySizeAndState = nil
	}
	if tsip.stoppedBrush != nil {
		tsip.stoppedBrush.Dispose()
		tsip.stoppedBrush = nil
	}
	if tsip.startingBrush != nil {
		tsip.startingBrush.Dispose()
		tsip.startingBrush = nil
	}
	if tsip.startedBrush != nil {
		tsip.startedBrush.Dispose()
		tsip.startedBrush = nil
	}
	if tsip.stoppedPen != nil {
		tsip.stoppedPen.Dispose()
		tsip.stoppedPen = nil
	}
	if tsip.startingPen != nil {
		tsip.startingPen.Dispose()
		tsip.startingPen = nil
	}
	if tsip.startedPen != nil {
		tsip.startedPen.Dispose()
		tsip.startedPen = nil
	}
}

func (tsip *TunnelStatusImageProvider) ImageForTunnel(tunnel *service.Tunnel, size walk.Size) (*walk.Bitmap, error) {
	state, err := tunnel.State()
	if err != nil {
		return nil, err
	}

	return tsip.ImageForState(state, size)
}

func (tsip *TunnelStatusImageProvider) ImageForState(state service.TunnelState, size walk.Size) (*walk.Bitmap, error) {
	key := sizeAndState{size, state}

	if img, ok := tsip.imagesBySizeAndState[key]; ok {
		return img, nil
	}

	var disposables walk.Disposables
	defer disposables.Treat()

	img, err := walk.NewBitmapWithTransparentPixels(size)
	if err != nil {
		return nil, err
	}

	canvas, err := walk.NewCanvasFromImage(img)
	if err != nil {
		return nil, err
	}
	defer canvas.Dispose()

	if err := tsip.PaintForState(state, canvas, walk.Rectangle{0, 0, size.Width, size.Height}); err != nil {
		return nil, err
	}

	tsip.imagesBySizeAndState[key] = img

	disposables.Spare()

	return img, nil
}

func (tsip *TunnelStatusImageProvider) PaintForTunnel(tunnel *service.Tunnel, canvas *walk.Canvas, bounds walk.Rectangle) error {
	state, err := tunnel.State()
	if err != nil {
		return err
	}

	return tsip.PaintForState(state, canvas, bounds)
}

func (tsip *TunnelStatusImageProvider) PaintForState(state service.TunnelState, canvas *walk.Canvas, bounds walk.Rectangle) error {
	var (
		brush *walk.SolidColorBrush
		pen   *walk.CosmeticPen
	)

	switch state {
	case service.TunnelStarted:
		brush = tsip.startedBrush
		pen = tsip.startedPen

	case service.TunnelStarting:
		brush = tsip.startingBrush
		pen = tsip.startingPen

	default:
		brush = tsip.stoppedBrush
		pen = tsip.stoppedPen
	}

	b := bounds

	b.X = 4
	b.Y += 4
	b.Height -= 8
	b.Width = b.Height

	if err := canvas.FillEllipse(brush, b); err != nil {
		return err
	}
	if err := canvas.DrawEllipse(pen, b); err != nil {
		return err
	}

	return nil
}
