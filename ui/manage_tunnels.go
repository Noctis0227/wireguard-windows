/* SPDX-License-Identifier: MIT
 *
 * Copyright (C) 2019 WireGuard LLC. All Rights Reserved.
 */

package ui

import (
	"archive/zip"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/lxn/walk"
	"github.com/lxn/win"
	"golang.zx2c4.com/wireguard/windows/conf"
	"golang.zx2c4.com/wireguard/windows/ringlogger"
	"golang.zx2c4.com/wireguard/windows/service"
)

type ManageTunnelsWindow struct {
	*walk.MainWindow

	icon *walk.Icon

	logger                 *ringlogger.Ringlogger
	tunnelTracker          *TunnelTracker
	tunnelsView            *TunnelsView
	confView               *ConfView
	tunnelAddedPublisher   walk.StringEventPublisher
	tunnelDeletedPublisher walk.StringEventPublisher
}

func NewManageTunnelsWindow(icon *walk.Icon, logger *ringlogger.Ringlogger) (*ManageTunnelsWindow, error) {
	var err error

	mtw := &ManageTunnelsWindow{
		icon:   icon,
		logger: logger,
	}
	mtw.MainWindow, err = walk.NewMainWindowWithName("WireGuard")
	if err != nil {
		return nil, err
	}

	return mtw, mtw.setup()
}

func (mtw *ManageTunnelsWindow) setup() error {
	mtw.SetIcon(mtw.icon)
	font, err := walk.NewFont("Segoe UI", 9, 0)
	if err != nil {
		return err
	}
	mtw.AddDisposable(font)
	mtw.SetFont(font)
	mtw.SetSize(walk.Size{900, 600})
	mtw.SetLayout(walk.NewVBoxLayout())
	mtw.Closing().Attach(func(canceled *bool, reason walk.CloseReason) {
		// "Close to tray" instead of exiting application
		onQuit()
	})

	splitter, _ := walk.NewHSplitter(mtw)
	splitter.SetSuspended(true)
	defer func() {
		splitter.SetSuspended(false)
	}()

	tunnelsContainer, _ := walk.NewComposite(splitter)
	tunnelsContainer.SetLayout(walk.NewVBoxLayout())

	splitter.SetFixed(tunnelsContainer, true)

	mtw.tunnelsView, _ = NewTunnelsView(tunnelsContainer)
	mtw.tunnelsView.ItemActivated().Attach(mtw.onEditTunnel)
	mtw.tunnelsView.CurrentIndexChanged().Attach(mtw.updateConfView)

	service.IPCClientRegisterTunnelChange(func(tunnel *service.Tunnel, state service.TunnelState, err error) {
		mtw.tunnelsView.Invalidate()
	})

	// ToolBar actions
	{
		// HACK: Because of https://github.com/lxn/walk/issues/481
		// we need to put the ToolBar into its own Composite.
		toolBarContainer, _ := walk.NewComposite(tunnelsContainer)
		toolBarContainer.SetLayout(walk.NewHBoxLayout())

		tunnelsToolBar, _ := walk.NewToolBar(toolBarContainer)

		importAction := walk.NewAction()
		importAction.SetText("Import tunnels from file...")
		importAction.Triggered().Attach(mtw.onImport)

		addAction := walk.NewAction()
		addAction.SetText("Add empty tunnel")
		addAction.Triggered().Attach(mtw.onAddTunnel)

		exportLogAction := walk.NewAction()
		exportLogAction.SetText("Export log to file...")
		exportLogAction.Triggered().Attach(mtw.onExportLog)

		exportTunnelsAction := walk.NewAction()
		exportTunnelsAction.SetText("Export tunnels to zip...")
		exportTunnelsAction.Triggered().Attach(mtw.onExportTunnels)

		// TODO: Add this to the dispose array (AddDisposable)
		addMenu, _ := walk.NewMenu()
		addMenu.Actions().Add(addAction)
		addMenu.Actions().Add(importAction)
		addMenuAction, _ := tunnelsToolBar.Actions().AddMenu(addMenu)
		addMenuAction.SetText("Add")

		deleteAction := walk.NewAction()
		tunnelsToolBar.Actions().Add(deleteAction)
		deleteAction.SetText("Delete")
		deleteAction.Triggered().Attach(mtw.onDelete)

		settingsMenu, _ := walk.NewMenu()
		settingsMenu.Actions().Add(exportLogAction)
		settingsMenu.Actions().Add(exportTunnelsAction)
		settingsMenuAction, _ := tunnelsToolBar.Actions().AddMenu(settingsMenu)
		settingsMenuAction.SetText("Export")
	}

	currentTunnelContainer, _ := walk.NewComposite(splitter)
	currentTunnelContainer.SetLayout(walk.NewVBoxLayout())

	mtw.confView, _ = NewConfView(currentTunnelContainer)
	go func() {
		// TODO: teardown in Dispose()
		t := time.NewTicker(time.Second)
		for range t.C {
			mtw.Synchronize(func() {
				mtw.updateConfView()
			})
		}
	}()

	// TODO: Find a better place for this?
	// logfile, err := service.IPCClientLogFilePath()
	// var logger *ringlogger.Ringlogger
	// if err == nil {
	// 	logger, err = ringlogger.NewRinglogger(logfile, "GUI")
	// }
	// if err != nil {
	// 	walk.MsgBox(nil, "Unable to initialize logging", fmt.Sprintf("%v\n\nFile: %s", err, logfile), walk.MsgBoxIconError)
	// 	return err
	// }
	// NewLogView(currentTunnelContainer, logger)

	controlsContainer, _ := walk.NewComposite(currentTunnelContainer)
	controlsContainer.SetLayout(walk.NewHBoxLayout())
	controlsContainer.Layout().SetMargins(walk.Margins{})

	walk.NewHSpacer(controlsContainer)

	editTunnel, _ := walk.NewPushButton(controlsContainer)
	editTunnel.SetEnabled(false)
	mtw.tunnelsView.CurrentIndexChanged().Attach(func() {
		editTunnel.SetEnabled(mtw.tunnelsView.CurrentIndex() > -1)
	})
	editTunnel.SetText("Edit")
	editTunnel.Clicked().Attach(mtw.onEditTunnel)

	return nil
}

func (mtw *ManageTunnelsWindow) Show() {
	mtw.MainWindow.Show()
	// TODO: Upstream lxn/walk has VisibleChanged()
	mtw.updateConfView()
	win.SetForegroundWindow(mtw.Handle())
	win.BringWindowToTop(mtw.Handle())
}

func (mtw *ManageTunnelsWindow) TunnelTracker() *TunnelTracker {
	return mtw.tunnelTracker
}

func (mtw *ManageTunnelsWindow) SetTunnelTracker(tunnelTracker *TunnelTracker) {
	mtw.tunnelTracker = tunnelTracker

	mtw.confView.SetTunnelTracker(tunnelTracker)
}

func (mtw *ManageTunnelsWindow) SetTunnelState(tunnel *service.Tunnel, state service.TunnelState) {
	mtw.tunnelsView.SetTunnelState(tunnel, state)
	// mtw.confView.SetTunnelState(tunnel, state)
}

func (mtw *ManageTunnelsWindow) updateConfView() {
	if !mtw.Visible() {
		return
	}

	mtw.confView.SetTunnel(mtw.tunnelsView.CurrentTunnel())
}

// importFiles tries to import a list of configurations.
func (mtw *ManageTunnelsWindow) importFiles(paths []string) {
	type unparsedConfig struct {
		Name   string
		Config string
	}

	var (
		unparsedConfigs []unparsedConfig
		lastErr         error
	)

	// Note: other versions of WireGuard start with all .zip files, then all .conf files.
	// To reproduce that if needed, inverse-sort the array.
	for _, path := range paths {
		switch filepath.Ext(path) {
		case ".conf":
			textConfig, err := ioutil.ReadFile(path)
			if err != nil {
				lastErr = err
				continue
			}
			unparsedConfigs = append(unparsedConfigs, unparsedConfig{Name: strings.TrimSuffix(filepath.Base(path), ".conf"), Config: string(textConfig)})
		case ".zip":
			// 1 .conf + 1 error .zip edge case?
			r, err := zip.OpenReader(path)
			if err != nil {
				lastErr = err
				continue
			}

			for _, f := range r.File {
				if filepath.Ext(f.Name) != ".conf" {
					continue
				}

				rc, err := f.Open()
				if err != nil {
					lastErr = err
					continue
				}
				textConfig, err := ioutil.ReadAll(rc)
				rc.Close()
				if err != nil {
					lastErr = err
					continue
				}
				unparsedConfigs = append(unparsedConfigs, unparsedConfig{Name: strings.TrimSuffix(filepath.Base(f.Name), ".conf"), Config: string(textConfig)})
			}

			r.Close()
		}
	}

	if lastErr != nil || unparsedConfigs == nil {
		walk.MsgBox(mtw, "Error", fmt.Sprintf("Could not parse some files: %v", lastErr), walk.MsgBoxIconWarning)
		return
	}

	var configs []*conf.Config

	for _, unparsedConfig := range unparsedConfigs {
		config, err := conf.FromWgQuick(unparsedConfig.Config, unparsedConfig.Name)
		if err != nil {
			lastErr = err
			continue
		}
		service.IPCClientNewTunnel(config)
		configs = append(configs, config)
	}

	m, n := len(configs), len(unparsedConfigs)
	switch {
	case n == 1 && m != n:
		walk.MsgBox(mtw, "Error", fmt.Sprintf("Could not parse some files: %v", lastErr), walk.MsgBoxIconWarning)
	case n == 1 && m == n:
		// TODO: Select tunnel in the list
	case m == n:
		walk.MsgBox(mtw, "Imported tunnels", fmt.Sprintf("Imported %d tunnels", m), walk.MsgBoxOK)
	case m != n:
		walk.MsgBox(mtw, "Imported tunnels", fmt.Sprintf("Imported %d of %d tunnels", m, n), walk.MsgBoxIconWarning)
	default:
		panic("unreachable case")
	}
}

func (mtw *ManageTunnelsWindow) exportTunnels(filePath string) {
	mtw.writeFileWithOverwriteHandling(filePath, func(file *os.File) error {
		writer := zip.NewWriter(file)

		for _, tunnel := range mtw.tunnelsView.model.tunnels {
			cfg, err := tunnel.StoredConfig()
			if err != nil {
				return fmt.Errorf("onExportTunnels: tunnel.StoredConfig failed: %v", err)
			}

			w, err := writer.Create(tunnel.Name + ".conf")
			if err != nil {
				return fmt.Errorf("onExportTunnels: writer.Create failed: %v", err)
			}

			if _, err := w.Write(([]byte)(cfg.ToWgQuick())); err != nil {
				return fmt.Errorf("onExportTunnels: cfg.ToWgQuick failed: %v", err)
			}
		}

		return writer.Close()
	})
}

func (mtw *ManageTunnelsWindow) addTunnel(config *conf.Config) {
	tunnel, err := service.IPCClientNewTunnel(config)
	if err != nil {
		walk.MsgBox(mtw, "Unable to create tunnel", err.Error(), walk.MsgBoxIconError)
		return
	}

	model := mtw.tunnelsView.model
	model.tunnels = append(model.tunnels, tunnel)
	model.PublishRowsReset()
	model.Sort(model.SortedColumn(), model.SortOrder())

	for i, t := range model.tunnels {
		if t.Name == tunnel.Name {
			mtw.tunnelsView.SetCurrentIndex(i)
			break
		}
	}

	mtw.confView.SetTunnel(&tunnel)

	mtw.tunnelAddedPublisher.Publish(tunnel.Name)
}

func (mtw *ManageTunnelsWindow) deleteTunnel(tunnel *service.Tunnel) {
	tunnel.Delete()

	model := mtw.tunnelsView.model

	for i, t := range model.tunnels {
		if t.Name == tunnel.Name {
			model.tunnels = append(model.tunnels[:i], model.tunnels[i+1:]...)
			model.PublishRowsRemoved(i, i)
			break
		}
	}

	mtw.tunnelDeletedPublisher.Publish(tunnel.Name)
}

func (mtw *ManageTunnelsWindow) TunnelAdded() *walk.StringEvent {
	return mtw.tunnelAddedPublisher.Event()
}

func (mtw *ManageTunnelsWindow) TunnelDeleted() *walk.StringEvent {
	return mtw.tunnelDeletedPublisher.Event()
}

func (mtw *ManageTunnelsWindow) exportLog(filePath string) {
	mtw.writeFileWithOverwriteHandling(filePath, func(file *os.File) error {
		if _, err := mtw.logger.WriteTo(file); err != nil {
			return fmt.Errorf("exportLog: Ringlogger.WriteTo failed: %v", err)
		}

		return nil
	})
}

func (mtw *ManageTunnelsWindow) writeFileWithOverwriteHandling(filePath string, write func(file *os.File) error) bool {
	showError := func(err error) bool {
		if err == nil {
			return false
		}

		walk.MsgBox(mtw, "Writing file failed", err.Error(), walk.MsgBoxIconError)

		return true
	}

	file, err := os.OpenFile(filePath, os.O_CREATE|os.O_WRONLY|os.O_EXCL, 0600)
	if err != nil {
		if os.IsExist(err) {
			if walk.DlgCmdNo == walk.MsgBox(mtw, "Writing file failed", fmt.Sprintf(`File "%s" already exists.

Do you want to overwrite it?`, filePath), walk.MsgBoxYesNo|walk.MsgBoxDefButton2|walk.MsgBoxIconWarning) {
				return false
			}

			if file, err = os.Create(filePath); err != nil {
				return !showError(err)
			}
		} else {
			return !showError(err)
		}
	}
	defer file.Close()

	return !showError(write(file))
}

// Handlers

func (mtw *ManageTunnelsWindow) onEditTunnel() {
	tunnel := mtw.tunnelsView.CurrentTunnel()
	if tunnel == nil {
		// Misfired event?
		return
	}

	if config := runTunnelConfigDialog(mtw, tunnel); config != nil {
		// Delete old one
		mtw.deleteTunnel(tunnel)

		// Save new one
		mtw.addTunnel(config)
	}
}

func (mtw *ManageTunnelsWindow) onAddTunnel() {
	if config := runTunnelConfigDialog(mtw, nil); config != nil {
		// Save new
		mtw.addTunnel(config)
	}
}

func (mtw *ManageTunnelsWindow) onDelete() {
	currentTunnel := mtw.tunnelsView.CurrentTunnel()
	if currentTunnel == nil {
		// Misfired event?
		return
	}

	if walk.DlgCmdNo == walk.MsgBox(
		mtw,
		fmt.Sprintf(`Delete "%s"`, currentTunnel.Name),
		fmt.Sprintf(`Are you sure you want to delete "%s"?`, currentTunnel.Name),
		walk.MsgBoxYesNo|walk.MsgBoxIconWarning) {
		return
	}

	mtw.deleteTunnel(currentTunnel)

	mtw.tunnelDeletedPublisher.Publish(currentTunnel.Name)
}

func (mtw *ManageTunnelsWindow) onImport() {
	dlg := &walk.FileDialog{}
	// dlg.InitialDirPath
	dlg.Filter = "Configuration Files (*.zip, *.conf)|*.zip;*.conf|All Files (*.*)|*.*"
	dlg.Title = "Import tunnel(s) from file..."

	if ok, _ := dlg.ShowOpenMultiple(mtw); !ok {
		return
	}

	mtw.importFiles(dlg.FilePaths)
}

func (mtw *ManageTunnelsWindow) onExportTunnels() {
	dlg := &walk.FileDialog{}
	dlg.Filter = "Configuration ZIP Files (*.zip)|*.zip"
	dlg.Title = "Export tunnels to zip..."

	if ok, _ := dlg.ShowSave(mtw); !ok {
		return
	}

	if !strings.HasSuffix(dlg.FilePath, ".zip") {
		dlg.FilePath += ".zip"
	}

	mtw.exportTunnels(dlg.FilePath)
}

func (mtw *ManageTunnelsWindow) onExportLog() {
	dlg := walk.FileDialog{
		Filter: "Log Files (*.log)|*.log|Text Files (*.txt)|*.txt|All Files (*.*)|*.*",
		Title:  "Export log to file",
	}

	if ok, _ := dlg.ShowSave(mtw); !ok {
		return
	}

	extensions := []string{".log", ".txt"}
	if dlg.FilterIndex < 3 && !strings.HasSuffix(dlg.FilePath, extensions[dlg.FilterIndex-1]) {
		dlg.FilePath = dlg.FilePath + extensions[dlg.FilterIndex-1]
	}

	mtw.exportLog(dlg.FilePath)
}
