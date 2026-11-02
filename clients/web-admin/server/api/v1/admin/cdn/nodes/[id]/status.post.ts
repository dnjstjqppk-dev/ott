import { defineEventHandler, readBody } from 'h3';

interface StatusBody {
  edge_node_id?: string;
  target_status?: 'HEALTHY' | 'DEGRADED' | 'DRAINING' | 'OFFLINE';
}

export default defineEventHandler(async (event) => {
  const body = await readBody<StatusBody>(event);
  const nodeId = event.context.params?.id || body?.edge_node_id || 'oca-icn-01';

  return {
    id: nodeId,
    node_code: 'KR-ICN-IX-01',
    region_code: 'ap-northeast-2',
    status: body?.target_status || 'HEALTHY',
    current_bandwidth_gbps: 84.6,
    cache_hit_ratio: 98.7,
    active_streams: 32410
  };
});
