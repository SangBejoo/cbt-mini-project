package history

import (
	base "cbt-test-mini-project/gen/proto"
	"cbt-test-mini-project/internal/entity"
	"cbt-test-mini-project/internal/usecase/history"
	"context"

	"google.golang.org/protobuf/types/known/timestamppb"
)

// historyHandler implements base.HistoryServiceServer
type historyHandler struct {
	base.UnimplementedHistoryServiceServer
	usecase history.HistoryUsecase
}

// NewHistoryHandler creates a new HistoryHandler
func NewHistoryHandler(usecase history.HistoryUsecase) base.HistoryServiceServer {
	return &historyHandler{usecase: usecase}
}

// GetStudentHistory gets student history
func (h *historyHandler) GetStudentHistory(ctx context.Context, req *base.StudentHistoryRequest) (*base.StudentHistoryResponse, error) {
	var tingkatan, idMataPelajaran *int
	if req.Tingkatan != 0 {
		t := int(req.Tingkatan)
		tingkatan = &t
	}
	if req.IdMataPelajaran != 0 {
		i := int(req.IdMataPelajaran)
		idMataPelajaran = &i
	}

	response, err := h.usecase.GetStudentHistory(req.NamaPeserta, tingkatan, idMataPelajaran, int(req.Pagination.Page), int(req.Pagination.PageSize))
	if err != nil {
		return nil, err
	}

	var histories []*base.HistorySummary
	for _, h := range response.History {
		var waktuSelesai *timestamppb.Timestamp
		if h.WaktuSelesai != nil {
			waktuSelesai = timestamppb.New(*h.WaktuSelesai)
		}

		histories = append(histories, &base.HistorySummary{
			Id:                    int32(h.ID),
			SessionToken:          h.SessionToken,
			MataPelajaran:         &base.MataPelajaran{Id: int32(h.MataPelajaran.ID), Nama: h.MataPelajaran.Nama},
			WaktuMulai:            timestamppb.New(h.WaktuMulai),
			WaktuSelesai:          waktuSelesai,
			DurasiPengerjaanDetik: int32(h.DurasiPengerjaanDetik),
			NilaiAkhir:            h.NilaiAkhir,
			JumlahBenar:           int32(h.JumlahBenar),
			TotalSoal:             int32(h.TotalSoal),
			Status:                base.TestStatus(base.TestStatus_value[string(h.Status)]),
		})
	}

	return &base.StudentHistoryResponse{
		NamaPeserta:       response.NamaPeserta,
		Tingkatan:         req.Tingkatan,
		History:           histories,
		RataRataNilai:     response.RataRataNilai,
		TotalTestCompleted: int32(response.TotalTestCompleted),
		Pagination: &base.PaginationResponse{
			TotalCount:  int32(response.Pagination.TotalCount),
			TotalPages:  int32(response.Pagination.TotalPages),
			CurrentPage: int32(response.Pagination.CurrentPage),
			PageSize:    int32(response.Pagination.PageSize),
		},
	}, nil
}

// GetHistoryDetail gets detailed history
func (h *historyHandler) GetHistoryDetail(ctx context.Context, req *base.GetHistoryDetailRequest) (*base.HistoryDetailResponse, error) {
	response, err := h.usecase.GetHistoryDetail(req.SessionToken)
	if err != nil {
		return nil, err
	}

	var jawabanDetails []*base.JawabanDetail
	for _, d := range response.DetailJawaban {
		var jawabanDipilih base.JawabanOption
		if d.JawabanDipilih != nil {
			jawabanDipilih = base.JawabanOption(base.JawabanOption_value[string(*d.JawabanDipilih)])
		}

		jawabanDetails = append(jawabanDetails, &base.JawabanDetail{
			NomorUrut:      int32(d.NomorUrut),
			Pertanyaan:     d.Pertanyaan,
			OpsiA:          d.OpsiA,
			OpsiB:          d.OpsiB,
			OpsiC:          d.OpsiC,
			OpsiD:          d.OpsiD,
			JawabanDipilih: jawabanDipilih,
			JawabanBenar:   base.JawabanOption(base.JawabanOption_value[string(d.JawabanBenar)]),
			IsCorrect:      d.IsCorrect,
		})
	}

	var breakdownMateri []*base.MateriBreakdown
	for _, b := range response.BreakdownMateri {
		breakdownMateri = append(breakdownMateri, &base.MateriBreakdown{
			NamaMateri:      b.NamaMateri,
			JumlahSoal:      int32(b.JumlahSoal),
			JumlahBenar:     int32(b.JumlahBenar),
			PersentaseBenar: b.PersentaseBenar,
		})
	}

	return &base.HistoryDetailResponse{
		SessionInfo:     h.convertToProtoTestSession(response.SessionInfo),
		DetailJawaban:   jawabanDetails,
		BreakdownMateri: breakdownMateri,
	}, nil
}

// Helper function to convert entity to proto
func (h *historyHandler) convertToProtoTestSession(session *entity.TestSession) *base.TestSession {
	var waktuSelesai, batasWaktu *timestamppb.Timestamp
	if session.WaktuSelesai != nil {
		waktuSelesai = timestamppb.New(*session.WaktuSelesai)
	}
	batasWaktu = timestamppb.New(session.BatasWaktu())

	var nilaiAkhir float64
	if session.NilaiAkhir != nil {
		nilaiAkhir = *session.NilaiAkhir
	}

	var jumlahBenar, totalSoal int32
	if session.JumlahBenar != nil {
		jumlahBenar = int32(*session.JumlahBenar)
	}
	if session.TotalSoal != nil {
		totalSoal = int32(*session.TotalSoal)
	}

	status := base.TestStatus(base.TestStatus_value[string(session.Status)])

	return &base.TestSession{
		Id:              int32(session.ID),
		SessionToken:    session.SessionToken,
		NamaPeserta:     session.NamaPeserta,
		Tingkatan:       int32(session.Tingkatan),
		MataPelajaran:   &base.MataPelajaran{Id: int32(session.MataPelajaran.ID), Nama: session.MataPelajaran.Nama},
		WaktuMulai:      timestamppb.New(session.WaktuMulai),
		WaktuSelesai:    waktuSelesai,
		BatasWaktu:      batasWaktu,
		DurasiMenit:     int32(session.DurasiMenit),
		NilaiAkhir:      nilaiAkhir,
		JumlahBenar:     jumlahBenar,
		TotalSoal:       totalSoal,
		Status:          status,
	}
}