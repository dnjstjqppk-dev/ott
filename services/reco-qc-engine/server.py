import os
import grpc
from concurrent import futures
import reco_pb2
import reco_pb2_grpc
from engine.lolomo_builder import LoLoMoEngine

class RecommendationServiceServicer(reco_pb2_grpc.RecommendationServiceServicer):
    def __init__(self):
        self.engine = LoLoMoEngine()

    def GenerateLoLoMo(self, request, context):
        data = self.engine.generate_lolomo(
            profile_id=request.profile_id,
            country_code=request.country_code,
            language_code=request.language_code,
            max_maturity=request.max_maturity
        )
        
        rows = []
        for r in data["rows"]:
            items = [
                reco_pb2.LoLoMoItemDto(
                    title_id=i["title_id"],
                    personalized_artwork_url=i["personalized_artwork_url"],
                    match_score_pct=i["match_score_pct"],
                    rank_number=i["rank_number"],
                    bookmark_ms=i["bookmark_ms"],
                    duration_ms=i["duration_ms"]
                ) for i in r["items"]
            ]
            rows.append(reco_pb2.LoLoMoRowDto(
                row_id=r["row_id"],
                row_index=r["row_index"],
                row_type=r["row_type"],
                header_text=r["header_text"],
                items=items
            ))

        hero = data["hero_billboard"]
        hero_dto = reco_pb2.LoLoMoItemDto(
            title_id=hero["title_id"],
            personalized_artwork_url=hero["personalized_artwork_url"],
            match_score_pct=hero["match_score_pct"],
            rank_number=hero["rank_number"]
        )

        return reco_pb2.GenerateLoLoMoResponse(
            snapshot_id=data["snapshot_id"],
            hero_billboard=hero_dto,
            rows=rows
        )

    def RecordBanditFeedback(self, request, context):
        self.engine.update_bandit(request.profile_id, request.artwork_id, request.event_type)
        return reco_pb2.RecordBanditFeedbackResponse(recorded=True)

def serve():
    addr = os.getenv("RECO_GRPC_ADDR", "0.0.0.0:50055")
    server = grpc.server(futures.ThreadPoolExecutor(max_workers=10))
    reco_pb2_grpc.add_RecommendationServiceServicer_to_server(RecommendationServiceServicer(), server)
    server.add_insecure_port(addr)
    server.start()
    print(f"[reco-qc-engine] python gRPC server listening on {addr}")
    server.wait_for_termination()

if __name__ == "__main__":
    serve()

