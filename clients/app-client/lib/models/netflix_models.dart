class LoLoMoItemModel {
  final String titleId;
  final String artworkUrl;
  final int matchScorePct;
  final int rankNumber;
  final int bookmarkMs;
  final int durationMs;

  const LoLoMoItemModel({
    required this.titleId,
    required this.artworkUrl,
    required this.matchScorePct,
    required this.rankNumber,
    required this.bookmarkMs,
    required this.durationMs,
  });

  factory LoLoMoItemModel.fromJson(Map<String, dynamic> json) {
    return LoLoMoItemModel(
      titleId: json['title_id'] as String? ?? '',
      artworkUrl: json['personalized_artwork_url'] as String? ?? '',
      matchScorePct: (json['match_score_pct'] as num?)?.toInt() ?? 95,
      rankNumber: (json['rank_number'] as num?)?.toInt() ?? 0,
      bookmarkMs: (json['bookmark_ms'] as num?)?.toInt() ?? 0,
      durationMs: (json['duration_ms'] as num?)?.toInt() ?? 1,
    );
  }
}

class LoLoMoRowModel {
  final String rowId;
  final String rowType;
  final String headerText;
  final List<LoLoMoItemModel> items;

  const LoLoMoRowModel({
    required this.rowId,
    required this.rowType,
    required this.headerText,
    required this.items,
  });

  factory LoLoMoRowModel.fromJson(Map<String, dynamic> json) {
    final rawItems = json['items'] as List<dynamic>? ?? const [];
    return LoLoMoRowModel(
      rowId: json['row_id'] as String? ?? '',
      rowType: json['row_type'] as String? ?? 'STANDARD',
      headerText: json['header_text'] as String? ?? '',
      items: rawItems
          .map((e) => LoLoMoItemModel.fromJson(e as Map<String, dynamic>))
          .toList(growable: false),
    );
  }
}

class PlaybackTicketModel {
  final String playbackSessionId;
  final String signedManifestUrl;
  final String drmLicenseToken;
  final String widevinePsshBase64;
  final int resumeBookmarkMs;
  final int introStartMs;
  final int introEndMs;

  const PlaybackTicketModel({
    required this.playbackSessionId,
    required this.signedManifestUrl,
    required this.drmLicenseToken,
    required this.widevinePsshBase64,
    required this.resumeBookmarkMs,
    required this.introStartMs,
    required this.introEndMs,
  });

  factory PlaybackTicketModel.fromJson(Map<String, dynamic> json) {
    return PlaybackTicketModel(
      playbackSessionId: json['playback_session_id'] as String,
      signedManifestUrl: json['signed_hls_manifest_url'] as String,
      drmLicenseToken: json['drm_license_token'] as String,
      widevinePsshBase64: json['widevine_pssh_base64'] as String? ?? '',
      resumeBookmarkMs: (json['resume_bookmark_ms'] as num?)?.toInt() ?? 0,
      introStartMs: (json['intro_start_ms'] as num?)?.toInt() ?? 15000,
      introEndMs: (json['intro_end_ms'] as num?)?.toInt() ?? 45000,
    );
  }
}

