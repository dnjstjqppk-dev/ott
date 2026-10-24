import 'package:dio/dio.dart';
import 'package:flutter/services.dart';
import '../models/netflix_models.dart';

class NetflixApiAndDrmBridge {
  static const MethodChannel _drmChannel =
      MethodChannel('com.netflix.clone/drm_offline');

  final Dio _dio;
  final String _profileId;

  NetflixApiAndDrmBridge({
    required String baseUrl,
    required String accessToken,
    required String profileId,
  })  : _profileId = profileId,
        _dio = Dio(
          BaseOptions(
            baseUrl: baseUrl,
            connectTimeout: const Duration(seconds: 8),
            receiveTimeout: const Duration(seconds: 8),
            headers: {
              'Authorization': 'Bearer $accessToken',
              'X-Netflix-Profile-Id': profileId,
            },
          ),
        );

  Future<List<LoLoMoRowModel>> fetchLoLoMoRows(String lang) async {
    final response = await _dio.get<Map<String, dynamic>>(
      '/api/v1/browse/lolomo',
      queryParameters: {'lang': lang},
    );
    final lolomo = response.data?['lolomo'] as Map<String, dynamic>? ?? {};
    final rows = lolomo['rows'] as List<dynamic>? ?? const [];
    return rows
        .map((r) => LoLoMoRowModel.fromJson(r as Map<String, dynamic>))
        .toList(growable: false);
  }

  Future<PlaybackTicketModel> acquirePlaybackTicket(String episodeId) async {
    final response = await _dio.post<Map<String, dynamic>>(
      '/api/v1/playback/session',
      data: {
        'profile_id': _profileId,
        'episode_id': episodeId,
      },
    );
    final ticket = PlaybackTicketModel.fromJson(response.data!);
    await _drmChannel.invokeMethod<void>('bindHardwareDrmSession', {
      'sessionId': ticket.playbackSessionId,
      'drmToken': ticket.drmLicenseToken,
      'psshBase64': ticket.widevinePsshBase64,
    });
    return ticket;
  }

  Future<bool> sendPlaybackHeartbeat({
    required String sessionId,
    required String episodeId,
    required int currentPosMs,
    required int watchedDeltaMs,
    required int bitrateKbps,
    required bool isPaused,
  }) async {
    final res = await _dio.post<Map<String, dynamic>>(
      '/api/v1/playback/heartbeat',
      data: {
        'playback_session_id': sessionId,
        'profile_id': _profileId,
        'episode_id': episodeId,
        'current_position_ms': currentPosMs,
        'watched_delta_ms': watchedDeltaMs,
        'current_bitrate_kbps': bitrateKbps,
        'buffer_underrun_delta': 0,
        'dropped_frames_delta': 0,
        'is_paused': isPaused,
      },
    );
    return res.data?['keep_alive_allowed'] as bool? ?? false;
  }

  Future<String> enqueueEncryptedOfflineDownload({
    required String episodeId,
    required String manifestUrl,
    required String drmToken,
    required bool wifiOnly,
  }) async {
    final localTaskId = await _drmChannel.invokeMethod<String>(
      'enqueueOfflineCmafDownload',
      {
        'episodeId': episodeId,
        'manifestUrl': manifestUrl,
        'drmToken': drmToken,
        'wifiOnly': wifiOnly,
        'expirationHours': 168,
      },
    );
    return localTaskId ?? '';
  }
}

