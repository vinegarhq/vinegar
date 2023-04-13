package main

import (
	"encoding/json"
	"log"
	"os"
	"path/filepath"
)

// Validate the given renderer, and apply it to the given map (fflags);
// It will also disable every other renderer.
func (c *Configuration) SetFFlagRenderer() {
	possibleRenderers := []string{
		"OpenGL",
		"D3D11FL10",
		"D3D11",
		"Vulkan",
	}

	validRenderer := false

	for _, r := range possibleRenderers {
		if c.Renderer == r {
			validRenderer = true
		}
	}

	if !validRenderer {
		log.Fatal("invalid renderer, must be one of:", possibleRenderers)
	}

	for _, r := range possibleRenderers {
		isRenderer := r == c.Renderer
		c.FFlags["FFlagDebugGraphicsPrefer"+r] = isRenderer
		c.FFlags["FFlagDebugGraphicsDisable"+r] = !isRenderer
	}
}

// Apply the configuration's FFlags to Roblox's FFlags file, named after app:
// ClientAppSettings.json, we also set (and check) the renderer specified in
// the configuration, then indent it to look pretty and write.
func (r *Roblox) ApplyFFlags(app string) {
	fflagsDir := filepath.Join(r.VersionDir, app+"Settings")
	CreateDirs(fflagsDir)

	fflagsFile, err := os.Create(filepath.Join(fflagsDir, app+"AppSettings.json"))
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("Applying custom FFlags to %s", app)

	Config.SetFFlagRenderer()

	fflagsJSON, err := json.MarshalIndent(Config.FFlags, "", "  ")
	if err != nil {
		log.Fatal(err)
	}

	if _, err := fflagsFile.Write(fflagsJSON); err != nil {
		log.Fatal(err)
	}
}

// 'Append' RCO's FFlags to the given map pointer.
func (c *Configuration) SetRCOFFlags() {
	log.Println("Applying RCO FFlags")
	// FIntFlagUpdateVersion: 78
	rco := map[string]interface{}{
		"FIntFlagUpdateVersion":                                         77,
		"FStringNewInGameMenuForcedUserIds":                             "2323539704;3822626535;4126963942;2627408279;4329723352",
		"DFIntSecondsBetweenDynamicVariableReloading":                   31557600,
		"FIntRolloutEnrollmentExpirationMinutes":                        31557600,
		"DFFlagDynamicFastVariableReloaderTest1":                        false,
		"FFlagExecDynInitTests":                                         false,
		"DFFlagEarlyExitIfHang":                                         false,
		"DFFlagRenderTC2_6":                                             true,
		"FFlagRenderTC2_Disable":                                        false,
		"FIntRenderTC2_PercentageRollout":                               100,
		"FFlagRenderFixTC2MemoryLeak":                                   true,
		"DFFlagRenderTC2PreferWidthPOT":                                 true,
		"DFFlagRenderTC2_DiscardGeometryData2":                          true,
		"FIntRenderTC2BakeQueueSize":                                    6,
		"FLogRenderTC2LogGroupAsync":                                    6,
		"FLogRenderTC2LogGroupContent":                                  6,
		"FLogRenderTC2LogGroupFlow":                                     6,
		"FLogRenderTC2LogGroupLOD":                                      true,
		"FFlagRenderTC2LogStateEnable":                                  true,
		"FFlagLoadCoreScriptsFromPatchOnly":                             true,
		"FFlagIncrementalPatchBuilderConfigurableRebuildPeriod":         true,
		"FIntMaxCachedPatches":                                          3,
		"DFFlagDynIpBlacklistBackoffEnabled":                            true,
		"DFIntAppLaunchFlowThrottlingRate":                              0,
		"FIntMeshContentProviderForceCacheSize":                         200000000,
		"FIntAnimationClipCacheBytes":                                   200000000,
		"FIntEmotesAnimationsPerPlayerCacheSize":                        200000000,
		"FIntDefaultMeshCacheSizeMB":                                    2048,
		"FIntSmoothTerrainPhysicsCacheSize":                             200000000,
		"DFIntHttpCurlConnectionCacheSize":                              200000000,
		"DFIntUserIdPlayerNameCacheSize":                                200000000,
		"DFIntUserIdPlayerNameLifetimeSeconds":                          604800,
		"FIntTaskSchedulerAutoThreadLimit":                              16,
		"DFIntTaskSchedulerTargetFps":                                   200000000,
		"FLogCSG3Details":                                               6,
		"FLogCSG3Errors":                                                6,
		"FLogCSG3Stats":                                                 6,
		"FLogCSG3Debug":                                                 6,
		"FLogCSGDetails":                                                6,
		"FLogCSGErrors":                                                 6,
		"FLogTextureManager":                                            6,
		"FLogTextureQualityLog":                                         6,
		"FLogGraphicsTextureReductionD3D11":                             6,
		"FLogGraphicsTextureReductionOpenGL":                            6,
		"FLogGraphicsTextureReductionVulkan":                            6,
		"FLogSSAO":                                                      6,
		"FIntSSAOMipLevels":                                             0,
		"DFIntJoinDataCompressionLevel":                                 8,
		"DFIntClusterCompressionLevel":                                  8,
		"DFIntNetworkSchemaCompressionRatio":                            18,
		"FFlagGraphicsGLTextureReduction":                               true,
		"FFlagHandleAltEnterFullscreenManually":                         false,
		"DFFlagTaskSchedulerAvoidYieldInBackground":                     true,
		"DFFlagTaskSchedulerBlockingShutdownInClients":                  true,
		"FFlagTaskSchedulerUseRobloxRuntime":                            true,
		"FFlagEnableQuickGameLaunch":                                    true,
		"FFlagNewEmotesInGame2":                                         true,
		"FFlagFlagAPICaching":                                           true,
		"FFlagPreloadAllFonts":                                          true,
		"FFlagPreloadMinimalFonts":                                      true,
		"FFlagRemoveRedundantFontPreloading":                            true,
		"FFlagPreloadTextureItemsOption4":                               true,
		"DFFlagContentProviderSupportsSymlinks":                         true,
		"DFFlagJointIrregularityOptimization":                           true,
		"DFFlagVariableDPIScale2":                                       true,
		"FFlagReduceGetFullNameAllocs":                                  true,
		"FFlagReduceGuiStateGfxGuiInvalidation":                         true,
		"FFlagAdornRenderDrawPolygonsAsUi":                              true,
		"DFFlagLogSystemSinkSupport2":                                   true,
		"DFFlagHttpClientOptimizeReqQueuing":                            true,
		"FFlagBatchAssetApi":                                            true,
		"DFFlagReportHttpBatchApiRejectedUrl":                           true,
		"DFLogBatchAssetApiLog":                                         6,
		"DFIntBatchThumbnailLimit":                                      200,
		"DFIntBatchThumbnailMaxReqests":                                 4,
		"DFIntBatchThumbnailResultsSizeCap":                             400,
		"DFIntHttpBatchApi_maxReqs":                                     16,
		"DFIntHttpBatchLimit":                                           256,
		"FFlagBatchGfxGui2":                                             true,
		"FFlagFinishFetchingAssetsCorrectly":                            true,
		"FFlagEnableZeroLatencyCacheForVersionedAssets":                 true,
		"FFlagAnimationClipMemCacheEnabled":                             true,
		"FFlagRenderEnableGlobalInstancing3":                            true,
		"FIntRenderEnableGlobalInstancingD3D10Percent":                  100,
		"FIntRenderEnableGlobalInstancingD3D11Percent":                  100,
		"FIntRenderEnableGlobalInstancingVulkanPercent":                 100,
		"FIntRenderEnableGlobalInstancingMetalPercent":                  100,
		"FFlagEnablePrefetchTimeout":                                    true,
		"DFFlagAddJobStartTimesExpiringPrefetch":                        true,
		"FFlagPrefetchOnEveryPlatform":                                  true,
		"FFlagCacheRequestedMaxSize":                                    true,
		"FFlagAudioAssetsInResizableCache2":                             true,
		"DFFlagCalculateKFSCachedDataEarly":                             true,
		"DFFlagEnablePerformanceControlSoundCache3":                     true,
		"FFlagFileSystemGetCacheDirectoryPointerCacheResult":            true,
		"DFFlagFileSystemGetCacheDirectoryLikeAndroid":                  true,
		"DFFlagHttpCacheCleanBasedOnMemory":                             true,
		"DFFlagHttpCacheMissingRedirects":                               true,
		"FFlagSimCSG3MegaAssetFetcherSkipCached":                        true,
		"FFlagSimCSGV3CacheVerboseBSPMemory":                            true,
		"FFlagSimCSGV3CacheVerboseCanary":                               true,
		"FFlagSimCSGV3CacheVerboseOperation":                            true,
		"FFlagSimCSGV3CacheVerboseSeparateDealloc":                      true,
		"FFlagSimUseCSGV3TreeRule2Reduction":                            true,
		"FFlagSimCSGV3SeparateDetachedCoplanarGraphs":                   true,
		"FFlagSimCSGV3PruneNonManifoldVertexGraph":                      true,
		"FFlagSimCSGV3MeshFromCollarNilFragment":                        true,
		"FFlagSimCSGV3isPlaneValidBugFix":                               true,
		"FFlagSimCSGV3InitializeVertexMapFix":                           true,
		"FFlagSimCSGV3IncrementalTriangulationStreamingCompression":     true,
		"FFlagSimCSGV3IncrementalTriangulationPhase3":                   true,
		"FFlagSimCSGV3IncrementalTriangulationPhase2":                   true,
		"FFlagSimCSGV3IncrementalTriangulationPhase1":                   true,
		"FFlagSimCSGV3EnableOperatorBodyPruning":                        true,
		"FFlagSimCSGV3AddIntersection":                                  true,
		"FFlagSimCSGKeepPhysicalConfigData":                             true,
		"FFlagSimCSGAllowLocalOperations":                               true,
		"DFFlagSimCSG3UseQuadBallInExperience":                          true,
		"FFlagSimCSG3UseMovedBuilders":                                  true,
		"FFlagSimCSG3NewAPIBreakApart":                                  true,
		"FFlagSimCSG3EnableNewAPI":                                      true,
		"FFlagSimCSG3AsyncWarmv2":                                       true,
		"FFlagSimCSG3AllowBodyToSave":                                   true,
		"FFlagCSGMeshDisableReadHash":                                   true,
		"FFlagCSGMeshDisableWriteHash":                                  true,
		"DFFlagCleanOldCSGData":                                         true,
		"FFlagWrapDispatcherTickLimit":                                  true,
		"FFlagGraphicsD3D11AllowThreadSafeTextureUpload":                true,
		"FFlagGraphicsDeviceEvents":                                     true,
		"FFlagGraphicsEnableD3D10Compute":                               true,
		"FFlagGraphicsTextureCopy":                                      true,
		"FIntRenderD3D11UncleanShutdownPercent":                         100,
		"FStringImmersiveAdsUniverseWhitelist":                          "0",
		"FFlagImmersiveAdsWhitelistDisabled":                            false,
		"FFlagAdGuiEnabled3":                                            false,
		"FFlagEnableAdsAPI":                                             false,
		"FFlagEnableBackendAdsProviderTimerService":                     false,
		"FFlagAdPortalEnabled3":                                         false,
		"FFlagAdServiceEnabled":                                         false,
		"FFlagEnableAdPortalTimerService2":                              false,
		"DFFlagEnableAdUnitName":                                        false,
		"FFlagEnableAdGuiTimerService":                                  false,
		"FFlagAdGuiTextLabelEnabled":                                    false,
		"FFlagAdPortalAdFetchEnabled":                                   false,
		"DFFlagAdGuiImpressionDisabled":                                 true,
		"FFlagFilteredLocalSimulation5":                                 true,
		"FFlagAllowHingedToAnchoredLocalSimulation":                     true,
		"DFIntLocalSimZonePercent":                                      50,
		"FFlagDisableOldCookieManagementSticky":                         true,
		"FFlagUnifiedCookieProtocolEnabledSticky":                       true,
		"DFFlagUnifiedCookieProtocolEnabled":                            true,
		"DFFlagAccessCookiesWithUrlEnabled":                             true,
		"FFlagAccessCookiesWithUrlEnabledSticky":                        true,
		"FIntEnableCullableScene2HundredthPercent":                      500,
		"DFFlagAudioUseVolumetricPanning":                               true,
		"DFFlagAudioVolumetricUtilsRefactor":                            true,
		"DFFlagAudioEnableVolumetricPanningForMeshes":                   true,
		"DFFlagAudioEnableVolumetricPanningForPolys":                    true,
		"DFFlagAlwaysPutSoundsOnDiskWhenLowOnMemory":                    true,
		"FFlagRemoveMemoryApsGpu":                                       true,
		"FFlagTrackAllDeviceMemory5":                                    true,
		"FIntAbuseReportScreenshotMaxSize":                              0,
		"DFIntCrashReportingHundredthsPercentage":                       0,
		"DFIntCrashUploadErrorInfluxHundredthsPercentage":               0,
		"DFIntCrashUploadToBacktracePercentage":                         0,
		"FFlagThreadCacheInit":                                          true,
		"FFlagUpdateUICachesWithQuadTree3":                              true,
		"DFFlagExperimentalRuntimeTextureCreation":                      true,
		"FFlagFixGraphicsQuality":                                       true,
		"FFlagCommitToGraphicsQualityFix":                               true,
		"FFlagFixTextureCompositorFramebufferManagement2":               true,
		"FFlagMemoryPrioritizationEnabledForTextures":                   true,
		"FFlagTextureManagerMaySkipBlackReloadFallback":                 true,
		"FFlagTextureManagerUsePerfControl":                             true,
		"FFlagTextureManagerUsePerfControlDirectMapping":                true,
		"FFlagTextureManagerUsePerfControlV2Api":                        true,
		"FFlagIntegrityCheckedProcessorUsePerfControl":                  true,
		"FFlagIntegrityCheckedProcessorPerfControlEffects":              true,
		"FFlagIntegrityCheckedProcessorUsePerfControlV2Api":             true,
		"FFlagPerfControlFireCallbacks2":                                true,
		"FFlagSoundServiceUsePerfControlV2Api":                          true,
		"FFlagPerformanceControlChangeTunableEagerly":                   true,
		"FFlagPerformanceControlDynamicUtilityCurves":                   true,
		"FFlagPerformanceControlMimicMemoryPrioritization":              true,
		"DFFlagPerformanceControlProportionalPlanner":                   true,
		"DFFlagPerformanceControlProportionalPlannerForV2":              true,
		"FFlagPerformanceControlSimpleMPLogic":                          true,
		"DFFlagESGamePerfMonitorEnabled":                                false,
		"DFIntESGamePerfMonitorHundredthsPercentage":                    0,
		"FIntGamePerfMonitorPercentage":                                 0,
		"DFFlagEnablePerfAudioCollection":                               false,
		"DFFlagEnablePerfDataCoreCategoryTimersCollection2":             false,
		"DFFlagEnablePerfDataCoreTimersCollection2":                     false,
		"DFFlagEnablePerfDataCountersCollection":                        false,
		"DFFlagEnablePerfDataGatherTelemetry2":                          false,
		"DFFlagEnablePerfDataMemoryCategoriesCollection2":               false,
		"DFFlagEnablePerfDataMemoryCollection":                          false,
		"DFFlagEnablePerfDataMemoryPerformanceCleanup3":                 false,
		"DFFlagEnablePerfDataMemoryPressureCollection":                  false,
		"DFFlagEnablePerfDataReportThermals":                            false,
		"DFFlagEnablePerfDataSubsystemTimersCollection2":                false,
		"DFFlagEnablePerfDataSummaryMode":                               false,
		"DFFlagEnablePerfRenderStatsCollection2":                        false,
		"DFFlagEnablePerfStatsCollection3":                              false,
		"FFlagRenderGpuTextureCompressor":                               true,
		"FFlagRenderLightGridEfficientTextureAtlasUpdate":               true,
		"FFlagSkipRenderIfDataModelBusy":                                true,
		"DFIntRenderingThrottleDelayInMS":                               100,
		"FFlagFontAtlasMipsAndRefactor":                                 true,
		"FFlagAddFontAtlasMipmaps":                                      true,
		"FFlagFixFontFamiliesNullCrash":                                 true,
		"FFlagReadHSRAlwaysVisibleData":                                 true,
		"FFlagApplyHSRAlwaysVisibleData":                                true,
		"FFlagLinearDeformerLocal":                                      true,
		"FFlagEnableLinearCageDeformer2":                                true,
		"FFlagHSRClusterImprovement":                                    true,
		"FFlagHSRRemoveDuplicateindices":                                true,
		"FFlagUseFallbackTextureStatusLoaded":                           true,
		"FFlagHumanoidDeferredSyncFunction5":                            true,
		"DFFlagHumanoidOnlyStepInWorkspace":                             true,
		"FFlagHumanoidParallelFasterSetCollision":                       true,
		"FFlagHumanoidParallelFasterWakeUp":                             true,
		"FFlagHumanoidParallelFixTickleFloor2":                          true,
		"FFlagHumanoidParallelOnStep2":                                  true,
		"FFlagHumanoidParallelSafeCofmUpdate":                           true,
		"FFlagHumanoidParallelSafeUnseat":                               true,
		"FFlagHumanoidParallelUseManager4":                              true,
		"FFlagCloudsUseBC4Compression":                                  true,
		"FIntClientCompressionFormatRequestPC":                          3,
		"FFlagCloudsMvpForceNoHistory":                                  true,
		"DFFlagThrottleDeveloperConsoleEvents":                          true,
		"FFlagFastGPULightCulling3":                                     true,
		"FFlagDebugForceFSMCPULightCulling":                             true,
		"DFFlagSimIfNoInterp2":                                          true,
		"DFFlagSimOptimizeInterpolationReturnPreviousIfSmallMovement2":  true,
		"FFlagMakeCSGPublishAsync":                                      true,
		"FFlagMakeHSRPublishAsync2":                                     true,
		"FFlagWriteFlagRolloutInformationToAttributesFile":              true,
		"DFFlagPlayerAvoidNetworkDependency":                            true,
		"FFlagUseMediumpSamplers":                                       true,
		"DFFlagClipMainAudioOutput":                                     true,
		"DFFlagCollectibleItemInExperiencePurchaseEnabled":              true,
		"DFFlagCollectibleItemInInspectAndBuyEnabled":                   true,
		"DFFlagFixUserInputServiceNotInitialized":                       true,
		"DFFlagIntegrateSendInPeer":                                     true,
		"DFFlagLocServiceGetSourceLanguageFromWebAPI":                   true,
		"DFFlagPlayerConfigurer2886":                                    true,
		"FFlagPlayerConfigurer2759":                                     true,
		"DFFlagPredictedOOM":                                            false,
		"DFFlagPredictedOOMAbs":                                         false,
		"DFFlagPredictedOOMMessageKeepPlayingLeave":                     false,
		"FFlagBatchThumbnailFetcherTimedOutSupport":                     true,
		"FFlagFileMeshToChunks2":                                        true,
		"FFlagKickClientOnCoreGuiRenderOverflow":                        false,
		"FFlagLaunchUAByDefault":                                        false,
		"FFlagUniversalAppOnUWP":                                        false,
		"FFlagUniversalAppOnWindows":                                    false,
		"FFlagEnableUniversalAppUserAgentOnMacAndWin32":                 false,
		"DFFlagEphemeralCounterInfluxReportingEnabled":                  false,
		"DFIntEphemeralCounterInfluxReportingPriorityHundredthsPercent": 0,
		"DFIntEphemeralCounterInfluxReportingThrottleHundredthsPercent": 100000,
		"DFIntEphemeralStatsInfluxReportingPriorityHundredthsPercent":   0,
		"DFIntEphemeralStatsInfluxReportingThrottleHundredthsPercent":   100000,
		"DFFlagDebugAnalyticsSendUserId":                                false,
		"DFStringAnalyticsEventStreamUrlEndpoint":                       "https://opt-out.roblox.com/%blackhole",
		"DFStringAltHttpPointsReporterUrl":                              "https://opt-out.roblox.com/%blackhole",
		"DFStringAltTelegrafHTTPTransportUrl":                           "https://opt-out.roblox.com/%blackhole",
		"DFStringTelegrafHTTPTransportUrl":                              "https://opt-out.roblox.com/%blackhole",
		"DFStringLmsRecipeEndpoint":                                     "/%blackhole",
		"DFStringLmsReportEndpoint":                                     "/%blackhole",
		"DFStringLightstepHTTPTransportUrlHost":                         "https://opt-out.roblox.com",
		"DFStringLightstepHTTPTransportUrlPath":                         "/%blackhole",
		"DFStringHttpPointsReporterUrl":                                 "https://opt-out.roblox.com/%blackhole",
		"DFStringCrashUploadToBacktraceBaseUrl":                         "https://opt-out.roblox.com/%blackhole",
		"DFStringRobloxAnalyticsSubDomain":                              "opt-out",
		"DFStringRobloxAnalyticsURL":                                    "https://opt-out.roblox.com/%blackhole",
		"DFFlagSendAllPhysicsPackets":                                   true,
	}

	for key, val := range c.FFlags {
		rco[key] = val
	}

	c.FFlags = rco
}
