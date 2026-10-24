import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import '../models/netflix_models.dart';
import '../services/api_and_drm_bridge.dart';
import 'native_player_screen.dart';

class TvAndMobileHomeScreen extends StatefulWidget {
  final NetflixApiAndDrmBridge bridge;

  const TvAndMobileHomeScreen({super.key, required this.bridge});

  @override
  State<TvAndMobileHomeScreen> createState() => _TvAndMobileHomeScreenState();
}

class _TvAndMobileHomeScreenState extends State<TvAndMobileHomeScreen> {
  List<LoLoMoRowModel> _rows = const [];
  bool _loading = true;
  String _focusedTitleId = '';

  @override
  void initState() {
    super.initState();
    _loadFeed();
  }

  Future<void> _loadFeed() async {
    final rows = await widget.bridge.fetchLoLoMoRows('ko');
    if (!mounted) return;
    setState(() {
      _rows = rows;
      _loading = false;
    });
  }

  Future<void> _startPlayback(String episodeOrTitleId) async {
    final ticket = await widget.bridge.acquirePlaybackTicket(episodeOrTitleId);
    if (!mounted) return;
    await Navigator.of(context).push(
      MaterialPageRoute<void>(
        builder: (_) => NativePlayerScreen(
          bridge: widget.bridge,
          episodeId: episodeOrTitleId,
          ticket: ticket,
        ),
      ),
    );
  }

  Future<void> _triggerSmartDownload(String episodeId) async {
    final ticket = await widget.bridge.acquirePlaybackTicket(episodeId);
    await widget.bridge.enqueueEncryptedOfflineDownload(
      episodeId: episodeId,
      manifestUrl: ticket.signedManifestUrl,
      drmToken: ticket.drmLicenseToken,
      wifiOnly: true,
    );
    if (!mounted) return;
    ScaffoldMessenger.of(context).showSnackBar(
      const SnackBar(content: Text('암호화된 오프라인 스마트 다운로드 큐에 등록되었습니다.')),
    );
  }

  @override
  Widget build(BuildContext context) {
    if (_loading) {
      return const Scaffold(
        backgroundColor: Colors.black,
        body: Center(child: CircularProgressIndicator(color: Colors.red)),
      );
    }
    return Scaffold(
      backgroundColor: Colors.black,
      appBar: AppBar(
        backgroundColor: Colors.black87,
        title: const Text(
          'NETFLIX',
          style: TextStyle(color: Colors.red, fontWeight: FontWeight.w900),
        ),
      ),
      body: FocusTraversalGroup(
        policy: ReadingOrderTraversalPolicy(),
        child: ListView.builder(
          itemCount: _rows.length,
          itemBuilder: (context, rowIdx) {
            final row = _rows[rowIdx];
            return Padding(
              padding: const EdgeInsets.symmetric(vertical: 12.0),
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  Padding(
                    padding: const EdgeInsets.symmetric(horizontal: 20.0),
                    child: Text(
                      row.headerText,
                      style: const TextStyle(
                        color: Colors.white,
                        fontSize: 18,
                        fontWeight: FontWeight.bold,
                      ),
                    ),
                  ),
                  const SizedBox(height: 8),
                  SizedBox(
                    height: 164,
                    child: ListView.builder(
                      scrollDirection: Axis.horizontal,
                      padding: const EdgeInsets.symmetric(horizontal: 16.0),
                      itemCount: row.items.length,
                      itemBuilder: (context, itemIdx) {
                        return _buildFocusableCard(row.items[itemIdx]);
                      },
                    ),
                  ),
                ],
              ),
            );
          },
        ),
      ),
    );
  }

  Widget _buildFocusableCard(LoLoMoItemModel item) {
    final isFocused = _focusedTitleId == item.titleId;
    return FocusableActionDetector(
      onShowFocusHighlight: (focused) {
        if (focused) setState(() => _focusedTitleId = item.titleId);
      },
      shortcuts: <LogicalKeySet, Intent>{
        LogicalKeySet(LogicalKeyboardKey.select): const ActivateIntent(),
        LogicalKeySet(LogicalKeyboardKey.enter): const ActivateIntent(),
      },
      actions: <Type, Action<Intent>>{
        ActivateIntent: CallbackAction<ActivateIntent>(
          onInvoke: (_) => _startPlayback(item.titleId),
        ),
      },
      child: GestureDetector(
        onTap: () => _startPlayback(item.titleId),
        onLongPress: () => _triggerSmartDownload(item.titleId),
        child: AnimatedContainer(
          duration: const Duration(milliseconds: 180),
          width: 220,
          margin: const EdgeInsets.symmetric(horizontal: 6.0),
          decoration: BoxDecoration(
            borderRadius: BorderRadius.circular(6),
            border: Border.all(
              color: isFocused ? Colors.white : Colors.transparent,
              width: 3,
            ),
          ),
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.stretch,
            children: [
              Expanded(
                child: Image.network(
                  item.artworkUrl,
                  fit: BoxFit.cover,
                ),
              ),
              if (item.bookmarkMs > 0 && item.durationMs > 0)
                LinearProgressIndicator(
                  value: (item.bookmarkMs / item.durationMs).clamp(0.0, 1.0),
                  color: Colors.red,
                  backgroundColor: Colors.grey.shade800,
                  minHeight: 4,
                ),
            ],
          ),
        ),
      ),
    );
  }
}

