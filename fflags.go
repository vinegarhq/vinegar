package main

import (
	"encoding/json"
	"log"
	"os"
	"path/filepath"
)

// Validate the given renderer, and apply it to the given map (fflags);
// It will also disable every other renderer.
func RobloxSetRenderer(renderer string, fflags *map[string]interface{}) {
	possibleRenderers := []string{
		"OpenGL",
		"D3D11FL10",
		"D3D11",
		"Vulkan",
	}

	validRenderer := false

	for _, r := range possibleRenderers {
		if renderer == r {
			validRenderer = true
		}
	}

	if !validRenderer {
		log.Fatal("invalid renderer, must be one of:", possibleRenderers)
	}

	for _, r := range possibleRenderers {
		isRenderer := r == renderer
		(*fflags)["FFlagDebugGraphicsPrefer"+r] = isRenderer
		(*fflags)["FFlagDebugGraphicsDisable"+r] = !isRenderer
	}
}

// Apply the configuration's FFlags to Roblox's FFlags file, named after app:
// ClientAppSettings.json, we also set (and check) the renderer specified in
// the configuration, then indent it to look pretty and write.
func RobloxApplyFFlags(app string, dir string) error {
	fflagsDir := filepath.Join(dir, app+"Settings")
	CreateDirs(fflagsDir)

	fflagsFile, err := os.Create(filepath.Join(fflagsDir, app+"AppSettings.json"))
	if err != nil {
		return err
	}

	log.Println("Applying custom FFlags")

	RobloxSetRenderer(Config.Renderer, &Config.FFlags)

	fflagsJSON, err := json.MarshalIndent(Config.FFlags, "", "  ")
	if err != nil {
		return err
	}

	if _, err := fflagsFile.Write(fflagsJSON); err != nil {
		return err
	}

	return nil
}

// 'Append' RCO's FFlags to the given map pointer.
func ApplyRCOFFlags(fflags *map[string]interface{}) {
	// FIntFlagUpdateVersion: 68
	rco := map[string]interface{}{
		"DFIntSecondsBetweenDynamicVariableReloading":                  31557600,
		"FIntRolloutEnrollmentExpirationMinutes":                       31557600,
		"DFFlagDynamicFastVariableReloaderTest1":                       false,
		"FIntMeshContentProviderForceCacheSize":                        200000000,
		"FIntAnimationClipCacheBytes":                                  200000000,
		"FIntEmotesAnimationsPerPlayerCacheSize":                       200000000,
		"FIntDefaultMeshCacheSizeMB":                                   2048,
		"FIntSmoothTerrainPhysicsCacheSize":                            200000000,
		"DFIntHttpCurlConnectionCacheSize":                             200000000,
		"DFIntUserIdPlayerNameCacheSize":                               200000000,
		"DFIntTaskSchedulerTargetFps":                                  200000000,
		"FIntSSAOMipLevels":                                            0,
		"DFIntExternalHttpRequestSizeLimitKB":                          4096,
		"DFIntJoinDataCompressionLevel":                                8,
		"DFIntClusterCompressionLevel":                                 8,
		"DFFlagRenderTC2_6":                                            false,
		"FFlagGraphicsGLTextureReduction":                              true,
		"FFlagHandleAltEnterFullscreenManually":                        false,
		"DFFlagTaskSchedulerAvoidYieldInBackground":                    true,
		"DFFlagTaskSchedulerBlockingShutdownInClients":                 true,
		"FFlagTaskSchedulerUseRobloxRuntime":                           true,
		"FFlagEnableQuickGameLaunch":                                   true,
		"FFlagPreloadAllFonts":                                         true,
		"FFlagPreloadMinimalFonts":                                     true,
		"FFlagRemoveRedundantFontPreloading":                           true,
		"FFlagPreloadTextureItemsOption4":                              true,
		"DFFlagContentProviderSupportsSymlinks":                        true,
		"DFFlagJointIrregularityOptimization":                          true,
		"DFFlagVariableDPIScale2":                                      true,
		"FFlagReduceGetFullNameAllocs":                                 true,
		"FFlagReduceGuiStateGfxGuiInvalidation":                        true,
		"FFlagAdornRenderDrawPolygonsAsUi":                             true,
		"DFFlagHttpClientOptimizeReqQueuing":                           true,
		"FFlagBatchAssetApi":                                           true,
		"DFFlagReportHttpBatchApiRejectedUrl":                          true,
		"DFIntBatchThumbnailLimit":                                     200,
		"DFIntBatchThumbnailMaxReqests":                                4,
		"DFIntBatchThumbnailResultsSizeCap":                            400,
		"DFIntHttpBatchApi_maxReqs":                                    16,
		"DFIntHttpBatchLimit":                                          256,
		"FFlagBatchGfxGui2":                                            true,
		"FFlagFinishFetchingAssetsCorrectly":                           true,
		"FFlagEnableZeroLatencyCacheForVersionedAssets":                true,
		"FFlagAnimationClipMemCacheEnabled":                            true,
		"FFlagRenderEnableGlobalInstancing3":                           true,
		"FIntRenderEnableGlobalInstancingD3D10Percent":                 100,
		"FIntRenderEnableGlobalInstancingVulkanPercent":                100,
		"FIntRenderEnableGlobalInstancingMetalPercent":                 100,
		"FFlagEnablePrefetchTimeout":                                   true,
		"DFFlagAddJobStartTimesExpiringPrefetch":                       true,
		"FFlagCacheRequestedMaxSize":                                   true,
		"FFlagAudioAssetsInResizableCache2":                            true,
		"DFFlagCalculateKFSCachedDataEarly":                            true,
		"DFFlagEnablePerformanceControlSoundCache3":                    true,
		"FFlagFileSystemGetCacheDirectoryPointerCacheResult":           true,
		"DFFlagFileSystemGetCacheDirectoryLikeAndroid":                 true,
		"DFFlagHttpCacheCleanBasedOnMemory":                            true,
		"DFFlagHttpCacheMissingRedirects":                              true,
		"FFlagSimCSG3MegaAssetFetcherSkipCached":                       true,
		"FFlagSimCSGV3CacheVerboseBSPMemory":                           true,
		"FFlagSimCSGV3CacheVerboseCanary":                              true,
		"FFlagSimCSGV3CacheVerboseOperation":                           true,
		"FFlagSimCSGV3CacheVerboseSeparateDealloc":                     true,
		"FFlagSimUseCSGV3TreeRule2Reduction":                           true,
		"FFlagSimCSGV3SeparateDetachedCoplanarGraphs":                  true,
		"FFlagSimCSGV3PruneNonManifoldVertexGraph":                     true,
		"FFlagSimCSGV3isPlaneValidBugFix":                              true,
		"FFlagSimCSGV3IncrementalTriangulationStreamingCompression":    true,
		"FFlagSimCSGV3IncrementalTriangulationPhase3":                  true,
		"FFlagSimCSGV3IncrementalTriangulationPhase2":                  true,
		"FFlagSimCSGV3IncrementalTriangulationPhase1":                  true,
		"FFlagSimCSGKeepPhysicalConfigData":                            true,
		"FFlagSimCSGAllowLocalOperations":                              true,
		"DFFlagSimCSG3UseQuadBallInExperience":                         true,
		"FFlagSimCSG3NewAPIBreakApart":                                 true,
		"FFlagSimCSG3EnableNewAPI":                                     true,
		"FFlagSimCSG3AsyncWarmv2":                                      true,
		"FFlagSimCSG3AllowBodyToSave":                                  true,
		"FFlagCSGMeshDisableReadHash":                                  true,
		"FFlagCSGMeshDisableWriteHash":                                 true,
		"DFFlagCleanOldCSGData":                                        true,
		"FFlagWrapDispatcherTickLimit":                                 true,
		"FFlagGraphicsD3D11AllowThreadSafeTextureUpload":               true,
		"FFlagGraphicsDeviceEvents":                                    true,
		"FFlagGraphicsEnableD3D10Compute":                              true,
		"FFlagGraphicsTextureCopy":                                     true,
		"FStringImmersiveAdsUniverseWhitelist":                         "0",
		"FFlagImmersiveAdsWhitelistDisabled":                           false,
		"FFlagAdGuiEnabled3":                                           false,
		"FFlagEnableAdsAPI":                                            false,
		"FFlagEnableBackendAdsProviderTimerService":                    false,
		"FFlagAdPortalEnabled3":                                        false,
		"FFlagAdServiceEnabled":                                        false,
		"FFlagEnableAdPortalTimerService2":                             false,
		"DFFlagEnableAdUnitName":                                       false,
		"FFlagEnableAdGuiTimerService":                                 false,
		"FFlagAdPortalAdFetchEnabled":                                  false,
		"DFFlagAdGuiImpressionDisabled":                                true,
		"FFlagFilteredLocalSimulation5":                                true,
		"FFlagAllowHingedToAnchoredLocalSimulation":                    true,
		"DFIntLocalSimZonePercent":                                     50,
		"FFlagDisableOldCookieManagementSticky":                        true,
		"FFlagUnifiedCookieProtocolEnabledSticky":                      true,
		"DFFlagUnifiedCookieProtocolEnabled":                           true,
		"DFFlagAccessCookiesWithUrlEnabled":                            true,
		"FFlagAccessCookiesWithUrlEnabledSticky":                       true,
		"FIntEnableCullableScene2HundredthPercent":                     500,
		"DFFlagAudioUseVolumetricPanning":                              true,
		"DFFlagAudioVolumetricUtilsRefactor":                           true,
		"DFFlagAudioEnableVolumetricPanningForMeshes":                  true,
		"DFFlagAudioEnableVolumetricPanningForPolys":                   true,
		"DFFlagAlwaysPutSoundsOnDiskWhenLowOnMemory":                   true,
		"FFlagRemoveMemoryApsGpu":                                      true,
		"FFlagTrackAllDeviceMemory5":                                   true,
		"FIntAbuseReportScreenshotMaxSize":                             0,
		"DFIntCrashReportingHundredthsPercentage":                      0,
		"DFIntCrashUploadErrorInfluxHundredthsPercentage":              0,
		"DFIntCrashUploadToBacktracePercentage":                        0,
		"FFlagThreadCacheInit":                                         true,
		"FFlagUpdateUICachesWithQuadTree3":                             true,
		"DFFlagExperimentalRuntimeTextureCreation":                     true,
		"FFlagFixTextureCompositorFramebufferManagement2":              true,
		"FFlagMemoryPrioritizationEnabledForTextures":                  true,
		"FFlagTextureManagerMaySkipBlackReloadFallback":                true,
		"FFlagTextureManagerUsePerfControl":                            true,
		"FFlagTextureManagerUsePerfControlDirectMapping":               true,
		"FFlagTextureManagerUsePerfControlV2Api":                       true,
		"FFlagIntegrityCheckedProcessorUsePerfControl":                 true,
		"FFlagIntegrityCheckedProcessorPerfControlEffects":             true,
		"FFlagIntegrityCheckedProcessorUsePerfControlV2Api":            true,
		"FFlagPerfControlFireCallbacks2":                               true,
		"FFlagSoundServiceUsePerfControlV2Api":                         true,
		"FFlagPerformanceControlChangeTunableEagerly":                  true,
		"FFlagPerformanceControlDynamicUtilityCurves":                  true,
		"FFlagPerformanceControlMimicMemoryPrioritization":             true,
		"DFFlagPerformanceControlProportionalPlanner":                  true,
		"DFFlagPerformanceControlProportionalPlannerForV2":             true,
		"FFlagPerformanceControlSimpleMPLogic":                         true,
		"DFFlagESGamePerfMonitorEnabled":                               false,
		"DFIntESGamePerfMonitorHundredthsPercentage":                   0,
		"FIntGamePerfMonitorPercentage":                                0,
		"DFFlagEnablePerfDataCoreCategoryTimersCollection2":            false,
		"DFFlagEnablePerfDataCoreTimersCollection2":                    false,
		"DFFlagEnablePerfDataGatherTelemetry2":                         false,
		"DFFlagEnablePerfDataMemoryCategoriesCollection2":              false,
		"DFFlagEnablePerfDataMemoryCollection":                         false,
		"DFFlagEnablePerfDataMemoryPressureCollection":                 false,
		"DFFlagEnablePerfDataSubsystemTimersCollection2":               false,
		"DFFlagEnablePerfDataSummaryMode":                              false,
		"DFFlagEnablePerfRenderStatsCollection2":                       false,
		"DFFlagEnablePerfStatsCollection3":                             false,
		"FFlagRenderGpuTextureCompressor":                              true,
		"FFlagRenderLightGridEfficientTextureAtlasUpdate":              true,
		"FFlagSkipRenderIfDataModelBusy":                               true,
		"DFIntRenderingThrottleDelayInMS":                              100,
		"FFlagFontAtlasMipsAndRefactor":                                true,
		"FFlagAddFontAtlasMipmaps":                                     true,
		"FFlagReadHSRAlwaysVisibleData":                                true,
		"FFlagApplyHSRAlwaysVisibleData":                               true,
		"FFlagLinearDeformerLocal":                                     true,
		"FFlagEnableLinearCageDeformer2":                               true,
		"FFlagHSRClusterImprovement":                                   true,
		"FFlagHSRRemoveDuplicateindices":                               true,
		"FFlagUseFallbackTextureStatusLoaded":                          true,
		"FFlagHumanoidDeferredSyncFunction5":                           true,
		"DFFlagHumanoidOnlyStepInWorkspace":                            true,
		"FFlagHumanoidParallelFasterSetCollision":                      true,
		"FFlagHumanoidParallelFasterWakeUp":                            true,
		"FFlagHumanoidParallelFixTickleFloor2":                         true,
		"FFlagHumanoidParallelOnStep2":                                 true,
		"FFlagHumanoidParallelSafeCofmUpdate":                          true,
		"FFlagHumanoidParallelSafeUnseat":                              true,
		"FFlagHumanoidParallelUseManager4":                             true,
		"FFlagCloudsUseBC4Compression":                                 true,
		"FIntClientCompressionFormatRequestPC":                         3,
		"FFlagCloudsMvpForceNoHistory":                                 true,
		"DFFlagThrottleDeveloperConsoleEvents":                         true,
		"FFlagFastGPULightCulling3":                                    true,
		"FFlagDebugForceFSMCPULightCulling":                            true,
		"DFFlagSimIfNoInterp2":                                         true,
		"DFFlagSimOptimizeInterpolationReturnPreviousIfSmallMovement2": true,
	}

	for key, val := range *fflags {
		rco[key] = val
	}

	(*fflags) = rco
}
