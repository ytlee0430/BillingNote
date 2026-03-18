package services

import (
	"billing-note/internal/models"
	"billing-note/internal/repository"
	"strings"

	"gorm.io/gorm"
)

type CategoryKeywordService struct {
	repo    repository.CategoryKeywordRepository
	catRepo repository.CategoryRepository
	db      *gorm.DB
}

func NewCategoryKeywordService(repo repository.CategoryKeywordRepository, catRepo repository.CategoryRepository, db *gorm.DB) *CategoryKeywordService {
	return &CategoryKeywordService{repo: repo, catRepo: catRepo, db: db}
}

// List returns all keyword rules for a user
func (s *CategoryKeywordService) List(userID uint) ([]models.CategoryKeyword, error) {
	return s.repo.ListByUser(userID)
}

// AddKeyword adds a single keyword rule
func (s *CategoryKeywordService) AddKeyword(userID, categoryID uint, keyword string) (*models.CategoryKeyword, error) {
	kw := &models.CategoryKeyword{
		UserID:     userID,
		CategoryID: categoryID,
		Keyword:    strings.TrimSpace(keyword),
	}
	if err := s.repo.Create(kw); err != nil {
		return nil, err
	}
	return kw, nil
}

// DeleteKeyword removes a keyword rule
func (s *CategoryKeywordService) DeleteKeyword(id, userID uint) error {
	return s.repo.Delete(id, userID)
}

// BatchSet replaces all keywords for a category
func (s *CategoryKeywordService) BatchSet(userID, categoryID uint, keywords []string) error {
	// Delete existing keywords for this category
	if err := s.repo.DeleteByUserAndCategory(userID, categoryID); err != nil {
		return err
	}

	// Create new ones
	var kws []models.CategoryKeyword
	for _, kw := range keywords {
		kw = strings.TrimSpace(kw)
		if kw == "" {
			continue
		}
		kws = append(kws, models.CategoryKeyword{
			UserID:     userID,
			CategoryID: categoryID,
			Keyword:    kw,
		})
	}
	return s.repo.BatchCreate(kws)
}

// ReclassifyAll applies keyword rules to all uncategorized transactions for a user.
// Returns the number of transactions updated.
func (s *CategoryKeywordService) ReclassifyAll(userID uint) (int, error) {
	keywords, err := s.repo.ListByUser(userID)
	if err != nil || len(keywords) == 0 {
		return 0, err
	}

	var transactions []models.Transaction
	if err := s.db.Where("user_id = ? AND category_id IS NULL", userID).Find(&transactions).Error; err != nil {
		return 0, err
	}

	updated := 0
	for _, t := range transactions {
		descNorm := normalizeForMatch(t.Description)
		for _, kw := range keywords {
			if strings.Contains(descNorm, normalizeForMatch(kw.Keyword)) {
				if err := s.db.Model(&t).Update("category_id", kw.CategoryID).Error; err == nil {
					updated++
				}
				break
			}
		}
	}

	return updated, nil
}

// MatchCategory finds the best category for a transaction description
func (s *CategoryKeywordService) MatchCategory(userID uint, description string) *uint {
	keywords, err := s.repo.ListByUser(userID)
	if err != nil || len(keywords) == 0 {
		return nil
	}

	descNorm := normalizeForMatch(description)
	for _, kw := range keywords {
		if strings.Contains(descNorm, normalizeForMatch(kw.Keyword)) {
			return &kw.CategoryID
		}
	}
	return nil
}

// normalizeForMatch converts fullwidth chars to halfwidth and lowercases for matching
func normalizeForMatch(s string) string {
	var b strings.Builder
	for _, r := range s {
		// Fullwidth ASCII (Ａ-Ｚ, ａ-ｚ, ０-９, etc.) → halfwidth
		if r >= 0xFF01 && r <= 0xFF5E {
			r = r - 0xFF01 + 0x0021
		}
		b.WriteRune(r)
	}
	return strings.ToLower(b.String())
}

// InitDefaults seeds default keyword rules for a user (called once)
func (s *CategoryKeywordService) InitDefaults(userID uint) error {
	existing, err := s.repo.ListByUser(userID)
	if err != nil {
		return err
	}
	if len(existing) > 0 {
		return nil // already initialized
	}

	categories, err := s.catRepo.GetAll()
	if err != nil {
		return err
	}

	catMap := make(map[string]uint)
	for _, c := range categories {
		catMap[c.Name] = c.ID
	}

	// Default keyword-to-category mappings
	defaults := map[string][]string{
		"餐飲": {
			"優食", "麥當勞", "漢堡王", "海底撈", "火鍋", "麵屋",
			"豆花", "飲茶", "subway", "鹹酥雞", "便當", "食品",
			"烘培", "餐廳", "美食", "泰滾", "朱記", "燒臘",
			"50嵐", "紅茶", "檸", "沾麵", "雞湯",
			"lawson", "seven-eleven", "全聯",
		},
		"交通": {
			"etag", "停車", "中油", "加油", "uber", "租車",
			"times car", "聯通", "監理",
		},
		"購物": {
			"好市多", "蝦皮", "taobao", "淘寶", "uniqlo",
			"寶雅", "poya", "屈臣氏", "大樹", "藥局",
			"家樂福", "宜得利", "百貨", "三越", "sogo",
			"全家", "統一超商", "連加", "連支",
		},
		"娛樂": {
			"遊戲",
		},
		"醫療": {
			"診所", "醫院", "安兒康",
		},
		"通訊": {
			"電話費", "中華電信",
		},
		"居家": {
			"電力", "瓦斯", "自來水", "台灣電力",
		},
		"訂閱": {
			"claude", "anthropic", "openai", "chatgpt",
			"google*cloud", "google*google one",
			"leetcode", "playstation network",
		},
		"其他支出": {
			"國外交易手續費", "年費",
		},
	}

	var kws []models.CategoryKeyword
	for catName, keywords := range defaults {
		catID, ok := catMap[catName]
		if !ok {
			continue
		}
		for _, kw := range keywords {
			kws = append(kws, models.CategoryKeyword{
				UserID:     userID,
				CategoryID: catID,
				Keyword:    kw,
			})
		}
	}

	return s.repo.BatchCreate(kws)
}
