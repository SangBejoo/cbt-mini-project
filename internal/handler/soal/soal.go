package soal

import (
	base "cbt-test-mini-project/gen/proto"
	"cbt-test-mini-project/internal/entity"
	"cbt-test-mini-project/internal/usecase/soal"
	"context"

	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// soalHandler implements base.SoalServiceServer
type soalHandler struct {
	base.UnimplementedSoalServiceServer
	usecase soal.SoalUsecase
}

// NewSoalHandler creates a new SoalHandler
func NewSoalHandler(usecase soal.SoalUsecase) base.SoalServiceServer {
	return &soalHandler{usecase: usecase}
}

// CreateSoal creates a new soal with multiple images
func (h *soalHandler) CreateSoal(ctx context.Context, req *base.CreateSoalRequest) (*base.SoalResponse, error) {
	jawabanBenar := entity.JawabanOption(req.JawabanBenar.String()[0])
	
	// Handle multiple image_bytes from repeated field
	var imageFilesBytes [][]byte
	if len(req.ImageBytes) > 0 {
		imageFilesBytes = req.ImageBytes
	}
	
	s, err := h.usecase.CreateSoal(int(req.IdMateri), int(req.IdTingkat), req.Pertanyaan, req.OpsiA, req.OpsiB, req.OpsiC, req.OpsiD, req.Pembahasan, jawabanBenar, imageFilesBytes)
	if err != nil {
		return nil, err
	}

	// Convert gambar entities to proto gambar
	var protoGambar []*base.SoalGambar
	for _, g := range s.Gambar {
		keteranganStr := ""
		if g.Keterangan != nil {
			keteranganStr = *g.Keterangan
		}
		cloudIdStr := ""
		if g.CloudId != nil {
			cloudIdStr = *g.CloudId
		}
		publicIdStr := ""
		if g.PublicId != nil {
			publicIdStr = *g.PublicId
		}
		
		protoGambar = append(protoGambar, &base.SoalGambar{
			Id:         int32(g.ID),
			NamaFile:   g.NamaFile,
			FilePath:   g.FilePath,
			FileSize:   int32(g.FileSize),
			MimeType:   g.MimeType,
			Urutan:     int32(g.Urutan),
			Keterangan: keteranganStr,
			CloudId:    cloudIdStr,
			PublicId:   publicIdStr,
			CreatedAt:  timestamppb.New(g.CreatedAt),
		})
	}

	return &base.SoalResponse{
		Soal: &base.SoalFull{
			Id: int32(s.ID),
			Materi: &base.Materi{
				Id:            int32(s.Materi.ID),
				MataPelajaran: &base.MataPelajaran{Id: int32(s.Materi.MataPelajaran.ID), Nama: s.Materi.MataPelajaran.Nama},
				Tingkat:       &base.Tingkat{Id: int32(s.Materi.Tingkat.ID), Nama: s.Materi.Tingkat.Nama},
				Nama:          s.Materi.Nama,
			},
			Pertanyaan:   s.Pertanyaan,
			OpsiA:        s.OpsiA,
			OpsiB:        s.OpsiB,
			OpsiC:        s.OpsiC,
			OpsiD:        s.OpsiD,
			JawabanBenar: base.JawabanOption(base.JawabanOption_value[string(s.JawabanBenar)]),
			Pembahasan: func() string {
				if s.Pembahasan != nil {
					return *s.Pembahasan
				}
				return ""
			}(),
			Gambar:       protoGambar,
		},
	}, nil
}

// GetSoal gets soal by ID
func (h *soalHandler) GetSoal(ctx context.Context, req *base.GetSoalRequest) (*base.SoalResponse, error) {
	s, err := h.usecase.GetSoal(int(req.Id))
	if err != nil {
		return nil, err
	}

	return &base.SoalResponse{
		Soal: &base.SoalFull{
			Id:            int32(s.ID),
			Materi:        &base.Materi{
				Id: int32(s.Materi.ID),
				MataPelajaran: &base.MataPelajaran{Id: int32(s.Materi.MataPelajaran.ID), Nama: s.Materi.MataPelajaran.Nama},
				Tingkat: &base.Tingkat{Id: int32(s.Materi.Tingkat.ID), Nama: s.Materi.Tingkat.Nama},
				Nama: s.Materi.Nama,
			},
			Pertanyaan:    s.Pertanyaan,
			OpsiA:         s.OpsiA,
			OpsiB:         s.OpsiB,
			OpsiC:         s.OpsiC,
			OpsiD:         s.OpsiD,
			JawabanBenar:  base.JawabanOption(base.JawabanOption_value[string(s.JawabanBenar)]),
			Pembahasan: func() string {
				if s.Pembahasan != nil {
					return *s.Pembahasan
				}
				return ""
			}(),
			Gambar:        convertSoalGambarToProto(s.Gambar),
		},
	}, nil
}

// UploadImageToSoal uploads an image to a soal
func (h *soalHandler) UploadImageToSoal(ctx context.Context, req *base.UploadImageToSoalRequest) (*base.UploadImageResponse, error) {
	var keterangan *string
	if req.Keterangan != "" {
		keterangan = &req.Keterangan
	}
	
	gambar, err := h.usecase.UploadImageToSoal(int(req.IdSoal), req.ImageBytes, req.NamaFile, int(req.Urutan), keterangan)
	if err != nil {
		return nil, err
	}

	keteranganStr := ""
	if gambar.Keterangan != nil {
		keteranganStr = *gambar.Keterangan
	}

	cloudIdStr := ""
	if gambar.CloudId != nil {
		cloudIdStr = *gambar.CloudId
	}

	publicIdStr := ""
	if gambar.PublicId != nil {
		publicIdStr = *gambar.PublicId
	}

	return &base.UploadImageResponse{
		Gambar: &base.SoalGambar{
			Id:         int32(gambar.ID),
			NamaFile:   gambar.NamaFile,
			FilePath:   gambar.FilePath,
			FileSize:   int32(gambar.FileSize),
			MimeType:   gambar.MimeType,
			Urutan:     int32(gambar.Urutan),
			Keterangan: keteranganStr,
			CloudId:    cloudIdStr,
			PublicId:   publicIdStr,
			CreatedAt:  timestamppb.New(gambar.CreatedAt),
		},
	}, nil
}

// DeleteImageFromSoal deletes an image from a soal
func (h *soalHandler) DeleteImageFromSoal(ctx context.Context, req *base.DeleteImageFromSoalRequest) (*base.MessageStatusResponse, error) {
	err := h.usecase.DeleteImageFromSoal(int(req.IdGambar))
	if err != nil {
		return nil, err
	}

	return &base.MessageStatusResponse{
		Message: "Image deleted successfully",
		Status:  "success",
	}, nil
}

// UpdateImageInSoal updates an image in a soal
func (h *soalHandler) UpdateImageInSoal(ctx context.Context, req *base.UpdateImageInSoalRequest) (*base.MessageStatusResponse, error) {
	var keterangan *string
	if req.Keterangan != "" {
		keterangan = &req.Keterangan
	}
	
	err := h.usecase.UpdateImageInSoal(int(req.IdGambar), int(req.Urutan), keterangan)
	if err != nil {
		return nil, err
	}

	return &base.MessageStatusResponse{
		Message: "Image updated successfully",
		Status:  "success",
	}, nil
}

// UpdateSoal updates soal with multiple images
func (h *soalHandler) UpdateSoal(ctx context.Context, req *base.UpdateSoalRequest) (*base.SoalResponse, error) {
	jawabanBenar := entity.JawabanOption(req.JawabanBenar.String()[0])
	
	// Handle multiple image_bytes from repeated field
	var imageFilesBytes [][]byte
	if len(req.ImageBytes) > 0 {
		imageFilesBytes = req.ImageBytes
	}
	
	s, err := h.usecase.UpdateSoal(int(req.Id), int(req.IdMateri), int(req.IdTingkat), req.Pertanyaan, req.OpsiA, req.OpsiB, req.OpsiC, req.OpsiD, req.Pembahasan, jawabanBenar, imageFilesBytes)
	if err != nil {
		return nil, err
	}

	// Convert gambar entities to proto gambar
	var protoGambar []*base.SoalGambar
	for _, g := range s.Gambar {
		protoGambar = append(protoGambar, &base.SoalGambar{
			Id:       int32(g.ID),
			FilePath: g.FilePath,
			Urutan:   int32(g.Urutan),
		})
	}

	return &base.SoalResponse{
		Soal: &base.SoalFull{
			Id: int32(s.ID),
			Materi: &base.Materi{
				Id:            int32(s.Materi.ID),
				MataPelajaran: &base.MataPelajaran{Id: int32(s.Materi.MataPelajaran.ID), Nama: s.Materi.MataPelajaran.Nama},
				Tingkat:       &base.Tingkat{Id: int32(s.Materi.Tingkat.ID), Nama: s.Materi.Tingkat.Nama},
				Nama:          s.Materi.Nama,
			},
			Pertanyaan:   s.Pertanyaan,
			OpsiA:        s.OpsiA,
			OpsiB:        s.OpsiB,
			OpsiC:        s.OpsiC,
			OpsiD:        s.OpsiD,
			JawabanBenar: base.JawabanOption(base.JawabanOption_value[string(s.JawabanBenar)]),
			Pembahasan: func() string {
				if s.Pembahasan != nil {
					return *s.Pembahasan
				}
				return ""
			}(),
			Gambar:       protoGambar,
		},
	}, nil
}

// DeleteSoal deletes soal
func (h *soalHandler) DeleteSoal(ctx context.Context, req *base.DeleteSoalRequest) (*base.MessageStatusResponse, error) {
	err := h.usecase.DeleteSoal(int(req.Id))
	if err != nil {
		return &base.MessageStatusResponse{
			Message: "Failed to delete soal",
			Status:  "error",
		}, err
	}

	return &base.MessageStatusResponse{
		Message: "Soal deleted successfully",
		Status:  "success",
	}, nil
}

// ListSoal lists soal
func (h *soalHandler) ListSoal(ctx context.Context, req *base.ListSoalRequest) (*base.ListSoalResponse, error) {
	page := 1
	pageSize := 1000 // Default to large number to get all for admin UI
	if req.Pagination != nil {
		if req.Pagination.Page > 0 {
			page = int(req.Pagination.Page)
		}
		if req.Pagination.PageSize > 0 {
			pageSize = int(req.Pagination.PageSize)
		}
	}
	soals, pagination, err := h.usecase.ListSoal(int(req.IdMateri), int(req.IdTingkat), int(req.IdMataPelajaran), page, pageSize)
	if err != nil {
		return nil, err
	}

	var soalList []*base.SoalFull
	for _, s := range soals {
		soalList = append(soalList, &base.SoalFull{
			Id:            int32(s.ID),
			Materi:        &base.Materi{
				Id: int32(s.Materi.ID),
				MataPelajaran: &base.MataPelajaran{Id: int32(s.Materi.MataPelajaran.ID), Nama: s.Materi.MataPelajaran.Nama},
				Tingkat: &base.Tingkat{Id: int32(s.Materi.Tingkat.ID), Nama: s.Materi.Tingkat.Nama},
				Nama: s.Materi.Nama,
			},
			Pertanyaan:    s.Pertanyaan,
			OpsiA:         s.OpsiA,
			OpsiB:         s.OpsiB,
			OpsiC:         s.OpsiC,
			OpsiD:         s.OpsiD,
			JawabanBenar:  base.JawabanOption(base.JawabanOption_value[string(s.JawabanBenar)]),
			Pembahasan: func() string {
				if s.Pembahasan != nil {
					return *s.Pembahasan
				}
				return ""
			}(),
			Gambar:         convertSoalGambarToProto(s.Gambar),
		})
	}

	return &base.ListSoalResponse{
		Soal: soalList,
		Pagination: &base.PaginationResponse{
			TotalCount:  int32(pagination.TotalCount),
			TotalPages:  int32(pagination.TotalPages),
			CurrentPage: int32(pagination.CurrentPage),
			PageSize:    int32(pagination.PageSize),
		},
	}, nil
}

// convertSoalGambarToProto converts entity.SoalGambar slice to proto SoalGambar slice
func convertSoalGambarToProto(gambar []entity.SoalGambar) []*base.SoalGambar {
	if len(gambar) == 0 {
		return nil
	}
	
	var protoGambar []*base.SoalGambar
	for _, g := range gambar {
		keteranganStr := ""
		if g.Keterangan != nil {
			keteranganStr = *g.Keterangan
		}
		
		cloudIdStr := ""
		if g.CloudId != nil {
			cloudIdStr = *g.CloudId
		}
		
		publicIdStr := ""
		if g.PublicId != nil {
			publicIdStr = *g.PublicId
		}
		
		protoGambar = append(protoGambar, &base.SoalGambar{
			Id:         int32(g.ID),
			NamaFile:   g.NamaFile,
			FilePath:   g.FilePath,
			FileSize:   int32(g.FileSize),
			MimeType:   g.MimeType,
			Urutan:     int32(g.Urutan),
			Keterangan: keteranganStr,
			CloudId:    cloudIdStr,
			PublicId:   publicIdStr,
			CreatedAt:  timestamppb.New(g.CreatedAt),
		})
	}
	return protoGambar
}

// GetQuestionCountsByTopic gets the count of questions per topic
func (h *soalHandler) GetQuestionCountsByTopic(ctx context.Context, req *emptypb.Empty) (*base.QuestionCountsResponse, error) {
	counts, err := h.usecase.GetQuestionCountsByTopic()
	if err != nil {
		return nil, err
	}

	var protoCounts []*base.TopicCount
	for topicId, count := range counts {
		protoCounts = append(protoCounts, &base.TopicCount{
			TopicId: int32(topicId),
			Count:   int32(count),
		})
	}

	return &base.QuestionCountsResponse{
		Counts: protoCounts,
	}, nil
}