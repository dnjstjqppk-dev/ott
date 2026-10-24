import 'dart:async';
import 'package:flutter/material.dart';
import 'package:video_player/video_player.dart';
import '../models/netflix_models.dart';
import '../services/api_and_drm_bridge.dart';

class NativePlayerScreen extends StatefulWidget {
  final NetflixApiAndDrmBridge bridge;
  final String episodeId;
  final PlaybackTicketModel ticket;

  const NativePlayerScreen({
    super.key,
    required this.bridge,
    required this.episodeId,
    required this.ticket,
  });

  @override
  State<NativePlayerScreen> createState() => _NativePlayerScreenState();
}

class _NativePlayerScreenState extends State<NativePlayerScreen> {
  late final VideoPlayerController _controller;
  Timer? _heartbeatTimer;
  int _lastReportedMs = 0;
  bool _showSkipIntro = false;

  @override
  void initState() {
    super.initState();
    _lastReportedMs = widget.ticket.resumeBookmarkMs;
    _controller = VideoPlayerController.networkUrl(
      Uri.parse(widget.ticket.signedManifestUrl),
    )..initialize().then((_) {
        _controller.seekTo(Duration(milliseconds: widget.ticket.resumeBookmarkMs));
        _controller.play();
        if (mounted) setState(() {});
      });
    _controller.addListener(_onPositionChanged);
    _startHeartbeatLoop();
  }

  void _onPositionChanged() {
    if (!_controller.value.isInitialized) return;
    final posMs = _controller.value.position.inMilliseconds;
    final inIntro = posMs >= widget.ticket.introStartMs &&
        posMs <= widget.ticket.introEndMs;
    if (inIntro != _showSkipIntro && mounted) {
      setState(() => _showSkipIntro = inIntro);
    }
  }

  void _startHeartbeatLoop() {
    _heartbeatTimer = Timer.periodic(const Duration(seconds: 5), (_) async {
      if (!_controller.value.isInitialized) return;
      final currentMs = _controller.value.position.inMilliseconds;
      final deltaMs = (currentMs - _lastReportedMs).clamp(0, 10000);
      _lastReportedMs = currentMs;

      final keepAlive = await widget.bridge.sendPlaybackHeartbeat(
        sessionId: widget.ticket.playbackSessionId,
        episodeId: widget.episodeId,
        currentPosMs: currentMs,
        watchedDeltaMs: deltaMs,
        bitrateKbps: 6500,
        isPaused: !_controller.value.isPlaying,
      );
      if (!keepAlive && mounted) {
        await _controller.pause();
        Navigator.of(context).pop();
      }
    });
  }

  void _skipIntro() {
    _controller.seekTo(Duration(milliseconds: widget.ticket.introEndMs));
  }

  @override
  void dispose() {
    _heartbeatTimer?.cancel();
    _controller.removeListener(_onPositionChanged);
    _controller.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      backgroundColor: Colors.black,
      body: Stack(
        children: [
          Center(
            child: _controller.value.isInitialized
                ? AspectRatio(
                    aspectRatio: _controller.value.aspectRatio,
                    child: VideoPlayer(_controller),
                  )
                : const CircularProgressIndicator(color: Colors.red),
          ),
          if (_showSkipIntro)
            Positioned(
              right: 32,
              bottom: 72,
              child: ElevatedButton(
                style: ElevatedButton.styleFrom(backgroundColor: Colors.white),
                onPressed: _skipIntro,
                child: const Text(
                  '오프닝 건너뛰기',
                  style: TextStyle(color: Colors.black, fontWeight: FontWeight.bold),
                ),
              ),
            ),
        ],
      ),
    );
  }
}

