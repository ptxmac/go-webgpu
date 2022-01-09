package wgpu

/*

#include <stdlib.h>

#include "./lib/webgpu.h"
#include "./lib/wgpu.h"

extern void requestAdapterCallback_cgo(WGPURequestAdapterStatus status,
                                WGPUAdapter adapter, char const *message,
                                void *userdata);

*/
import "C"

import (
	"errors"
	"runtime/cgo"
	"unsafe"
)

type AdapterExtras struct {
	BackendType BackendType
}

type RequestAdapterOptions struct {
	CompatibleSurface *Surface
	PowerPreference   PowerPreference

	// ChainedStruct -> WGPUAdapterExtras
	AdapterExtras *AdapterExtras

	// unused in wgpu
	// ForceFallbackAdapter bool
}

type requestAdapterCB func(status RequestAdapterStatus, adapter *Adapter, message string)

func RequestAdapter(options RequestAdapterOptions) (*Adapter, error) {
	var opts C.WGPURequestAdapterOptions

	if options.CompatibleSurface != nil {
		opts.compatibleSurface = options.CompatibleSurface.ref
	}
	opts.powerPreference = C.WGPUPowerPreference(options.PowerPreference)

	if options.AdapterExtras != nil {
		adapterExtras := (*C.WGPUAdapterExtras)(C.malloc(C.size_t(unsafe.Sizeof(C.WGPUAdapterExtras{}))))
		defer C.free(unsafe.Pointer(adapterExtras))

		adapterExtras.chain.next = nil
		adapterExtras.chain.sType = C.WGPUSType_AdapterExtras
		adapterExtras.backend = C.WGPUBackendType(options.AdapterExtras.BackendType)

		opts.nextInChain = (*C.WGPUChainedStruct)(unsafe.Pointer(adapterExtras))
	}

	var status RequestAdapterStatus
	var adapter *Adapter

	var cb requestAdapterCB = func(s RequestAdapterStatus, a *Adapter, _ string) {
		status = s
		adapter = a
	}
	handle := cgo.NewHandle(cb)
	C.wgpuInstanceRequestAdapter(nil, &opts, C.WGPURequestAdapterCallback(C.requestAdapterCallback_cgo), unsafe.Pointer(&handle))

	if status != RequestAdapterStatus_Success {
		return nil, errors.New("failed to request adapter")
	}
	return adapter, nil
}

type SurfaceDescriptorFromWindowsHWND struct {
	Hinstance unsafe.Pointer
	Hwnd      unsafe.Pointer
}

type SurfaceDescriptorFromXlib struct {
	Display unsafe.Pointer
	Window  uint32
}

type SurfaceDescriptorFromMetalLayer struct {
	Layer unsafe.Pointer
}

type SurfaceDescriptor struct {
	Label string

	// ChainedStruct -> WGPUSurfaceDescriptorFromWindowsHWND
	WindowsHWND *SurfaceDescriptorFromWindowsHWND

	// ChainedStruct -> WGPUSurfaceDescriptorFromXlib
	Xlib *SurfaceDescriptorFromXlib

	// ChainedStruct -> WGPUSurfaceDescriptorFromMetalLayer
	MetalLayer *SurfaceDescriptorFromMetalLayer
}

func CreateSurface(descriptor SurfaceDescriptor) *Surface {
	var desc C.WGPUSurfaceDescriptor

	if descriptor.Label != "" {
		label := C.CString(descriptor.Label)
		defer C.free(unsafe.Pointer(label))

		desc.label = label
	}

	if descriptor.WindowsHWND != nil {
		windowsHWND := (*C.WGPUSurfaceDescriptorFromWindowsHWND)(C.malloc(C.size_t(unsafe.Sizeof(C.WGPUSurfaceDescriptorFromWindowsHWND{}))))
		defer C.free(unsafe.Pointer(windowsHWND))

		windowsHWND.chain.next = nil
		windowsHWND.chain.sType = C.WGPUSType_SurfaceDescriptorFromWindowsHWND
		windowsHWND.hinstance = descriptor.WindowsHWND.Hinstance
		windowsHWND.hwnd = descriptor.WindowsHWND.Hwnd

		desc.nextInChain = (*C.WGPUChainedStruct)(unsafe.Pointer(windowsHWND))
	}

	if descriptor.Xlib != nil {
		xlib := (*C.WGPUSurfaceDescriptorFromXlib)(C.malloc(C.size_t(unsafe.Sizeof(C.WGPUSurfaceDescriptorFromXlib{}))))
		defer C.free(unsafe.Pointer(xlib))

		xlib.chain.next = nil
		xlib.chain.sType = C.WGPUSType_SurfaceDescriptorFromXlib
		xlib.display = descriptor.Xlib.Display
		xlib.window = C.uint32_t(descriptor.Xlib.Window)

		desc.nextInChain = (*C.WGPUChainedStruct)(unsafe.Pointer(xlib))
	}

	if descriptor.MetalLayer != nil {
		metalLayer := (*C.WGPUSurfaceDescriptorFromMetalLayer)(C.malloc(C.size_t(unsafe.Sizeof(C.WGPUSurfaceDescriptorFromMetalLayer{}))))
		defer C.free(unsafe.Pointer(metalLayer))

		metalLayer.chain.next = nil
		metalLayer.chain.sType = C.WGPUSType_SurfaceDescriptorFromMetalLayer
		metalLayer.layer = descriptor.MetalLayer.Layer

		desc.nextInChain = (*C.WGPUChainedStruct)(unsafe.Pointer(metalLayer))
	}

	ref := C.wgpuInstanceCreateSurface(nil, &desc)
	if ref == nil {
		return nil
	}
	return &Surface{ref}
}
