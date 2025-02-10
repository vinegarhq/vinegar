/*
layer to fix VK_ERROR_OUT_OF_DATE_KHR errors in roblox studio
written by tunis
workaround by V3L0C1T13S

based on https://github.com/baldurk/sample_layer/
license of the sample layer:
BSD 2-Clause License

Copyright (c) 2016, Baldur Karlsson
All rights reserved.

Redistribution and use in source and binary forms, with or without
modification, are permitted provided that the following conditions are met:

* Redistributions of source code must retain the above copyright notice, this
  list of conditions and the following disclaimer.

* Redistributions in binary form must reproduce the above copyright notice,
  this list of conditions and the following disclaimer in the documentation
  and/or other materials provided with the distribution.

THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS "AS IS"
AND ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT LIMITED TO, THE
IMPLIED WARRANTIES OF MERCHANTABILITY AND FITNESS FOR A PARTICULAR PURPOSE ARE
DISCLAIMED. IN NO EVENT SHALL THE COPYRIGHT HOLDER OR CONTRIBUTORS BE LIABLE
FOR ANY DIRECT, INDIRECT, INCIDENTAL, SPECIAL, EXEMPLARY, OR CONSEQUENTIAL
DAMAGES (INCLUDING, BUT NOT LIMITED TO, PROCUREMENT OF SUBSTITUTE GOODS OR
SERVICES; LOSS OF USE, DATA, OR PROFITS; OR BUSINESS INTERRUPTION) HOWEVER
CAUSED AND ON ANY THEORY OF LIABILITY, WHETHER IN CONTRACT, STRICT LIABILITY,
OR TORT (INCLUDING NEGLIGENCE OR OTHERWISE) ARISING IN ANY WAY OUT OF THE USE
OF THIS SOFTWARE, EVEN IF ADVISED OF THE POSSIBILITY OF SUCH DAMAGE.
*/

#include "vk_layer.h"

#include <assert.h>
#include <string.h>

#include <cstdio>
#include <mutex>
#include <map>

#undef VK_LAYER_EXPORT
#if defined(WIN32)
#define VK_LAYER_EXPORT extern "C" __declspec(dllexport)
#else
#define VK_LAYER_EXPORT extern "C"
#endif

// single global lock, for simplicity
std::mutex global_lock;
typedef std::lock_guard<std::mutex> scoped_lock;

// use the loader's dispatch table pointer as a key for dispatch map lookups
template<typename DispatchableType>
void *GetKey(DispatchableType inst)
{
    return *(void **)inst;
}

// layer book-keeping information, to store dispatch tables by key
std::map<void *, VkLayerInstanceDispatchTable> instance_dispatch;
std::map<void *, VkLayerDispatchTable> device_dispatch;

// actual data we're recording in this layer
bool hack_swapchain_recreation = false;

///////////////////////////////////////////////////////////////////////////////////////////
// Layer init and shutdown

VK_LAYER_EXPORT VkResult VKAPI_CALL VinegarLayer_CreateInstance(
        const VkInstanceCreateInfo*                 pCreateInfo,
        const VkAllocationCallbacks*                pAllocator,
        VkInstance*                                 pInstance)
{
    VkLayerInstanceCreateInfo *layerCreateInfo = (VkLayerInstanceCreateInfo *)pCreateInfo->pNext;

    // step through the chain of pNext until we get to the link info
    while(layerCreateInfo && (layerCreateInfo->sType != VK_STRUCTURE_TYPE_LOADER_INSTANCE_CREATE_INFO ||
                                                        layerCreateInfo->function != VK_LAYER_LINK_INFO))
    {
        layerCreateInfo = (VkLayerInstanceCreateInfo *)layerCreateInfo->pNext;
    }

    if(layerCreateInfo == NULL)
    {
        // No loader instance create info
        return VK_ERROR_INITIALIZATION_FAILED;
    }

    PFN_vkGetInstanceProcAddr gpa = layerCreateInfo->u.pLayerInfo->pfnNextGetInstanceProcAddr;
    // move chain on for next layer
    layerCreateInfo->u.pLayerInfo = layerCreateInfo->u.pLayerInfo->pNext;

    PFN_vkCreateInstance createFunc = (PFN_vkCreateInstance)gpa(VK_NULL_HANDLE, "vkCreateInstance");

    VkResult ret = createFunc(pCreateInfo, pAllocator, pInstance);

    // fetch our own dispatch table for the functions we need, into the next layer
    VkLayerInstanceDispatchTable dispatchTable;
    dispatchTable.GetInstanceProcAddr = (PFN_vkGetInstanceProcAddr)gpa(*pInstance, "vkGetInstanceProcAddr");
    dispatchTable.DestroyInstance = (PFN_vkDestroyInstance)gpa(*pInstance, "vkDestroyInstance");
    dispatchTable.EnumerateDeviceExtensionProperties = (PFN_vkEnumerateDeviceExtensionProperties)gpa(*pInstance, "vkEnumerateDeviceExtensionProperties");
    dispatchTable.GetPhysicalDeviceSurfaceCapabilitiesKHR = (PFN_vkGetPhysicalDeviceSurfaceCapabilitiesKHR)gpa(*pInstance, "vkGetPhysicalDeviceSurfaceCapabilitiesKHR");

    // store the table by key
    {
        scoped_lock l(global_lock);
        instance_dispatch[GetKey(*pInstance)] = dispatchTable;
    }

    return VK_SUCCESS;
}

VK_LAYER_EXPORT void VKAPI_CALL VinegarLayer_DestroyInstance(VkInstance instance, const VkAllocationCallbacks* pAllocator)
{
    scoped_lock l(global_lock);
    instance_dispatch.erase(GetKey(instance));
}

VK_LAYER_EXPORT VkResult VKAPI_CALL VinegarLayer_CreateDevice(
        VkPhysicalDevice                            physicalDevice,
        const VkDeviceCreateInfo*                   pCreateInfo,
        const VkAllocationCallbacks*                pAllocator,
        VkDevice*                                   pDevice)
{
    VkLayerDeviceCreateInfo *layerCreateInfo = (VkLayerDeviceCreateInfo *)pCreateInfo->pNext;

    // step through the chain of pNext until we get to the link info
    while(layerCreateInfo && (layerCreateInfo->sType != VK_STRUCTURE_TYPE_LOADER_DEVICE_CREATE_INFO ||
                                                        layerCreateInfo->function != VK_LAYER_LINK_INFO))
    {
        layerCreateInfo = (VkLayerDeviceCreateInfo *)layerCreateInfo->pNext;
    }

    if(layerCreateInfo == NULL)
    {
        // No loader instance create info
        return VK_ERROR_INITIALIZATION_FAILED;
    }
    
    PFN_vkGetInstanceProcAddr gipa = layerCreateInfo->u.pLayerInfo->pfnNextGetInstanceProcAddr;
    PFN_vkGetDeviceProcAddr gdpa = layerCreateInfo->u.pLayerInfo->pfnNextGetDeviceProcAddr;
    // move chain on for next layer
    layerCreateInfo->u.pLayerInfo = layerCreateInfo->u.pLayerInfo->pNext;

    PFN_vkCreateDevice createFunc = (PFN_vkCreateDevice)gipa(VK_NULL_HANDLE, "vkCreateDevice");

    VkResult ret = createFunc(physicalDevice, pCreateInfo, pAllocator, pDevice);
    
    // fetch our own dispatch table for the functions we need, into the next layer
    VkLayerDispatchTable dispatchTable;
    dispatchTable.GetDeviceProcAddr = (PFN_vkGetDeviceProcAddr)gdpa(*pDevice, "vkGetDeviceProcAddr");
    dispatchTable.DestroyDevice = (PFN_vkDestroyDevice)gdpa(*pDevice, "vkDestroyDevice");
    dispatchTable.AcquireNextImageKHR = (PFN_vkAcquireNextImageKHR)gdpa(*pDevice, "vkAcquireNextImageKHR");
    
    // store the table by key
    {
        scoped_lock l(global_lock);
        device_dispatch[GetKey(*pDevice)] = dispatchTable;
    }

    return VK_SUCCESS;
}

VK_LAYER_EXPORT void VKAPI_CALL VinegarLayer_DestroyDevice(VkDevice device, const VkAllocationCallbacks* pAllocator)
{
    scoped_lock l(global_lock);
    device_dispatch.erase(GetKey(device));
}

///////////////////////////////////////////////////////////////////////////////////////////
// Actual layer implementation

VK_LAYER_EXPORT VkResult VKAPI_CALL VinegarLayer_GetPhysicalDeviceSurfaceCapabilitiesKHR(VkPhysicalDevice physicalDevice, VkSurfaceKHR surface, VkSurfaceCapabilitiesKHR *pSurfaceCapabilities)
{
    scoped_lock l(global_lock);
    VkResult result = instance_dispatch[GetKey(physicalDevice)].GetPhysicalDeviceSurfaceCapabilitiesKHR(physicalDevice, surface, pSurfaceCapabilities);

    // this "fixes" VK_ERROR_OUT_OF_DATE_KHR and VK_SUBOPTIMAL_KHR by taking
    // advantage of roblox recreating the swapchain when we return
    // VK_ERROR_SURFACE_LOST_KHR
    if (hack_swapchain_recreation) {
        hack_swapchain_recreation = false;
        return VK_ERROR_SURFACE_LOST_KHR;
    }
    return result;
}

VK_LAYER_EXPORT VkResult VKAPI_CALL VinegarLayer_AcquireNextImageKHR(VkDevice device, VkSwapchainKHR swapchain, uint64_t timeout, VkSemaphore semaphore, VkFence fence, uint32_t *pImageIndex)
{
    scoped_lock l(global_lock);
    VkResult result = device_dispatch[GetKey(device)].AcquireNextImageKHR(device, swapchain, timeout, semaphore, fence, pImageIndex);

    // this error code is basically a non-fatal VK_ERROR_OUT_OF_DATE_KHR,
    // it reports that the surface properties have changed but the swapchain
    // is still usable. roblox cant recognise this, so we have to convert it
    if (result == VK_SUBOPTIMAL_KHR) {
        hack_swapchain_recreation = true;
        return VK_SUCCESS;
    } else if (result == VK_ERROR_OUT_OF_DATE_KHR)
        hack_swapchain_recreation = true;

    return result;
}

///////////////////////////////////////////////////////////////////////////////////////////
// Enumeration function

VK_LAYER_EXPORT VkResult VKAPI_CALL VinegarLayer_EnumerateInstanceLayerProperties(uint32_t *pPropertyCount,
                                                                                                                                             VkLayerProperties *pProperties)
{
    if(pPropertyCount) *pPropertyCount = 1;

    if(pProperties)
    {
        strcpy(pProperties->layerName, "VK_LAYER_VINEGAR_VinegarLayer");
        strcpy(pProperties->description, "Vinegar layer");
        pProperties->implementationVersion = 1;
        pProperties->specVersion = VK_API_VERSION_1_0;
    }

    return VK_SUCCESS;
}

VK_LAYER_EXPORT VkResult VKAPI_CALL VinegarLayer_EnumerateDeviceLayerProperties(
        VkPhysicalDevice physicalDevice, uint32_t *pPropertyCount, VkLayerProperties *pProperties)
{
    return VinegarLayer_EnumerateInstanceLayerProperties(pPropertyCount, pProperties);
}

VK_LAYER_EXPORT VkResult VKAPI_CALL VinegarLayer_EnumerateInstanceExtensionProperties(
        const char *pLayerName, uint32_t *pPropertyCount, VkExtensionProperties *pProperties)
{
    if(pLayerName == NULL || strcmp(pLayerName, "VK_LAYER_VINEGAR_VinegarLayer"))
        return VK_ERROR_LAYER_NOT_PRESENT;

    // don't expose any extensions
    if(pPropertyCount) *pPropertyCount = 0;
    return VK_SUCCESS;
}

VK_LAYER_EXPORT VkResult VKAPI_CALL VinegarLayer_EnumerateDeviceExtensionProperties(
                                                                         VkPhysicalDevice physicalDevice, const char *pLayerName,
                                                                         uint32_t *pPropertyCount, VkExtensionProperties *pProperties)
{
    // pass through any queries that aren't to us
    if(pLayerName == NULL || strcmp(pLayerName, "VK_LAYER_VINEGAR_VinegarLayer"))
    {
        if(physicalDevice == VK_NULL_HANDLE)
            return VK_SUCCESS;

        scoped_lock l(global_lock);
        return instance_dispatch[GetKey(physicalDevice)].EnumerateDeviceExtensionProperties(physicalDevice, pLayerName, pPropertyCount, pProperties);
    }

    // don't expose any extensions
    if(pPropertyCount) *pPropertyCount = 0;
    return VK_SUCCESS;
}

///////////////////////////////////////////////////////////////////////////////////////////
// GetProcAddr functions, entry points of the layer

#define GETPROCADDR(func) if(!strcmp(pName, "vk" #func)) return (PFN_vkVoidFunction)&VinegarLayer_##func;

VK_LAYER_EXPORT PFN_vkVoidFunction VKAPI_CALL VinegarLayer_GetDeviceProcAddr(VkDevice device, const char *pName)
{
    // device chain functions we intercept
    GETPROCADDR(GetDeviceProcAddr);
    GETPROCADDR(EnumerateDeviceLayerProperties);
    GETPROCADDR(EnumerateDeviceExtensionProperties);
    GETPROCADDR(CreateDevice);
    GETPROCADDR(DestroyDevice);
    GETPROCADDR(AcquireNextImageKHR);
    
    {
        scoped_lock l(global_lock);
        return device_dispatch[GetKey(device)].GetDeviceProcAddr(device, pName);
    }
}

VK_LAYER_EXPORT PFN_vkVoidFunction VKAPI_CALL VinegarLayer_GetInstanceProcAddr(VkInstance instance, const char *pName)
{
    // instance chain functions we intercept
    GETPROCADDR(GetInstanceProcAddr);
    GETPROCADDR(EnumerateInstanceLayerProperties);
    GETPROCADDR(EnumerateInstanceExtensionProperties);
    GETPROCADDR(CreateInstance);
    GETPROCADDR(DestroyInstance);
    GETPROCADDR(GetPhysicalDeviceSurfaceCapabilitiesKHR);
    
    // device chain functions we intercept
    GETPROCADDR(GetDeviceProcAddr);
    GETPROCADDR(EnumerateDeviceLayerProperties);
    GETPROCADDR(EnumerateDeviceExtensionProperties);
    GETPROCADDR(CreateDevice);
    GETPROCADDR(DestroyDevice);
    GETPROCADDR(AcquireNextImageKHR);

    {
        scoped_lock l(global_lock);
        return instance_dispatch[GetKey(instance)].GetInstanceProcAddr(instance, pName);
    }
}
