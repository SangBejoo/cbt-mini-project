package test_session

import (
	base "cbt-test-mini-project/gen/proto"
	"cbt-test-mini-project/internal/entity"
	"cbt-test-mini-project/internal/usecase/test_session"
	tingkatUsecase "cbt-test-mini-project/internal/usecase/tingkat"
	"context"
	"errors"
	"strings"

	"google.golang.org/protobuf/types/known/timestamppb"
)

// testSessionHandler implements base.TestSessionServiceServer
type testSessionHandler struct {
	base.UnimplementedTestSessionServiceServer
	usecase        test_session.TestSessionUsecase
	tingkatUsecase tingkatUsecase.TingkatUsecase
}

// NewTestSessionHandler creates a new TestSessionHandler
func NewTestSessionHandler(usecase test_session.TestSessionUsecase, tingkatUsecase tingkatUsecase.TingkatUsecase) base.TestSessionServiceServer {
	return &testSessionHandler{usecase: usecase, tingkatUsecase: tingkatUsecase}
}

// CreateTestSession creates a new test session
func (h *testSessionHandler) CreateTestSession(ctx context.Context, req *base.CreateTestSessionRequest) (*base.TestSessionResponse, error) {
	session, err := h.usecase.CreateTestSession(req.NamaPeserta, int(req.IdTingkat), int(req.IdMataPelajaran), int(req.DurasiMenit), int(req.JumlahSoal))
	if err != nil {
		return nil, err
	}

	return &base.TestSessionResponse{
		TestSession: h.convertToProtoTestSession(session),
	}, nil
}

// GetTestSession gets session by token
func (h *testSessionHandler) GetTestSession(ctx context.Context, req *base.GetTestSessionRequest) (*base.TestSessionResponse, error) {
	session, err := h.usecase.GetTestSession(req.SessionToken)
	if err != nil {
		return nil, err
	}

	return &base.TestSessionResponse{
		TestSession: h.convertToProtoTestSession(session),
	}, nil
}

// GetTestQuestions gets all questions for the session
func (h *testSessionHandler) GetTestQuestions(ctx context.Context, req *base.GetTestQuestionsRequest) (*base.TestQuestionsResponse, error) {
	session, err := h.usecase.GetTestSession(req.SessionToken)
	if err != nil {
		return nil, err
	}
	if session == nil {
		return nil, errors.New("session not found")
	}

	soals, err := h.usecase.GetAllTestQuestions(req.SessionToken)
	if err != nil {
		return nil, err
	}

	// Get answers status
	answers, _ := h.usecase.GetSessionAnswers(req.SessionToken)
	isAnsweredStatus := make([]bool, len(soals))
	for i := range soals {
		for _, ans := range answers {
			if ans.TestSessionSoal.NomorUrut == soals[i].NomorUrut {
				isAnsweredStatus[i] = true
				break
			}
		}
	}

	var protoSoals []*base.SoalForStudent
	for _, s := range soals {
		var jawabanDipilih base.JawabanOption
		if s.JawabanDipilih != nil {
			jawabanDipilih = base.JawabanOption(base.JawabanOption_value[string(*s.JawabanDipilih)])
		}

		protoSoals = append(protoSoals, &base.SoalForStudent{
			Id:             int32(s.ID),
			NomorUrut:      int32(s.NomorUrut),
			Pertanyaan:     s.Pertanyaan,
			OpsiA:          s.OpsiA,
			OpsiB:          s.OpsiB,
			OpsiC:          s.OpsiC,
			OpsiD:          s.OpsiD,
			JawabanDipilih: jawabanDipilih,
			IsAnswered:     s.IsAnswered,
			Materi: &base.Materi{
				Id:             int32(s.Materi.ID),
				Nama:           s.Materi.Nama,
				MataPelajaran:  &base.MataPelajaran{Id: int32(s.Materi.MataPelajaran.ID), Nama: s.Materi.MataPelajaran.Nama},
				Tingkat:        &base.Tingkat{Id: int32(s.Materi.Tingkat.ID), Nama: s.Materi.Tingkat.Nama},
			},
			Gambar:         convertSoalGambarToProto(s.Gambar),
		})
	}

	return &base.TestQuestionsResponse{
		SessionToken:      req.SessionToken,
		Soal:              protoSoals,
		TotalSoal:         int32(len(protoSoals)),
		CurrentNomorUrut:  1, // Not used
		DijawabCount:      int32(len(answers)),
		IsAnsweredStatus:  isAnsweredStatus,
		BatasWaktu:        timestamppb.New(session.BatasWaktu()),
	}, nil
}

// SubmitAnswer submits an answer
func (h *testSessionHandler) SubmitAnswer(ctx context.Context, req *base.SubmitAnswerRequest) (*base.SubmitAnswerResponse, error) {
	jawaban := entity.JawabanOption(req.JawabanDipilih.String()[0])
	err := h.usecase.SubmitAnswer(req.SessionToken, int(req.NomorUrut), jawaban)
	if err != nil {
		return nil, err
	}

	return &base.SubmitAnswerResponse{
		SessionToken:    req.SessionToken,
		NomorUrut:       req.NomorUrut,
		JawabanDipilih:  req.JawabanDipilih,
		IsCorrect:       true, // TODO: get from usecase
		DijawabPada:     timestamppb.Now(),
	}, nil
}

// CompleteSession completes the session
func (h *testSessionHandler) CompleteSession(ctx context.Context, req *base.CompleteSessionRequest) (*base.TestSessionResponse, error) {
	session, err := h.usecase.CompleteSession(req.SessionToken)
	if err != nil {
		return nil, err
	}

	return &base.TestSessionResponse{
		TestSession: h.convertToProtoTestSession(session),
	}, nil
}

// GetTestResult gets test result
func (h *testSessionHandler) GetTestResult(ctx context.Context, req *base.GetTestResultRequest) (*base.TestResultResponse, error) {
	session, details, err := h.usecase.GetTestResult(req.SessionToken)
	if err != nil {
		return nil, err
	}

	// Get all tingkat
	tingkatList, _, err := h.tingkatUsecase.ListTingkat(1, 100) // Assuming max 100 tingkat
	if err != nil {
		return nil, err
	}

	var jawabanDetails []*base.JawabanDetail
	for _, d := range details {
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

	var protoTingkat []*base.Tingkat
	for _, t := range tingkatList {
		protoTingkat = append(protoTingkat, &base.Tingkat{
			Id:   int32(t.ID),
			Nama: t.Nama,
		})
	}

	return &base.TestResultResponse{
		SessionInfo:   h.convertToProtoTestSession(session),
		DetailJawaban: jawabanDetails,
		Tingkat:       protoTingkat,
	}, nil
}

// ListTestSessions lists sessions
func (h *testSessionHandler) ListTestSessions(ctx context.Context, req *base.ListTestSessionsRequest) (*base.ListTestSessionsResponse, error) {
	var tingkatan, idMataPelajaran *int
	var status *entity.TestStatus

	if req.IdTingkat != 0 {
		t := int(req.IdTingkat)
		tingkatan = &t
	}
	if req.IdMataPelajaran != 0 {
		i := int(req.IdMataPelajaran)
		idMataPelajaran = &i
	}
	if req.Status != base.TestStatus_STATUS_INVALID {
		s := entity.TestStatus(req.Status.String())
		status = &s
	}

	page := 1
	pageSize := 10
	if req.Pagination != nil {
		if req.Pagination.Page > 0 {
			page = int(req.Pagination.Page)
		}
		if req.Pagination.PageSize > 0 {
			pageSize = int(req.Pagination.PageSize)
		}
	}

	sessions, pagination, err := h.usecase.ListTestSessions(tingkatan, idMataPelajaran, status, page, pageSize)
	if err != nil {
		return nil, err
	}

	var sessionList []*base.TestSession
	for _, s := range sessions {
		sessionList = append(sessionList, h.convertToProtoTestSession(&s))
	}

	return &base.ListTestSessionsResponse{
		TestSessions: sessionList,
		Pagination: &base.PaginationResponse{
			TotalCount:  int32(pagination.TotalCount),
			TotalPages:  int32(pagination.TotalPages),
			CurrentPage: int32(pagination.CurrentPage),
			PageSize:    int32(pagination.PageSize),
		},
	}, nil
}

// Helper function to convert entity to proto
func (h *testSessionHandler) convertToProtoTestSession(session *entity.TestSession) *base.TestSession {
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

	status := base.TestStatus(base.TestStatus_value[strings.ToUpper(string(session.Status))])

	return &base.TestSession{
		Id:              int32(session.ID),
		SessionToken:    session.SessionToken,
		NamaPeserta:     session.NamaPeserta,
		Tingkat:         &base.Tingkat{Id: int32(session.Tingkat.ID), Nama: session.Tingkat.Nama},
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

// convertSoalGambarToProto converts entity.SoalGambar slice to proto SoalGambar slice
func convertSoalGambarToProto(gambar []entity.SoalGambar) []*base.SoalGambar {
	if len(gambar) == 0 {
		return nil
	}
	
	var protoGambar []*base.SoalGambar
	for _, g := range gambar {
		protoGambar = append(protoGambar, &base.SoalGambar{
			Id:       int32(g.ID),
			FilePath: g.FilePath,
			Urutan:   int32(g.Urutan),
		})
	}
	return protoGambar
}