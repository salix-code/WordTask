package mock

// Word 单词数据结构（Mock 阶段，字段与后续 DB 模型保持一致）
type Word struct {
	ID             int      `json:"id"`
	Term           string   `json:"term"`
	Phonetic       string   `json:"phonetic"`
	PartOfSpeech   string   `json:"partOfSpeech"`
	Translation    string   `json:"translation"`
	Examples       []string `json:"examples"`
	Theme          string   `json:"theme"`
	Priority       string   `json:"priority"`
	LearningTarget string   `json:"learningTarget"`
}

// TodayWords 模拟今日待复习的单词列表
var TodayWords = []Word{
	{
		ID:             1,
		Term:           "hello",
		Phonetic:       "/həˈloʊ/",
		PartOfSpeech:   "int",
		Translation:    "你好；喂",
		Examples:       []string{"Hello, how are you?", "Say hello to your mother."},
		Theme:          "greetings",
		Priority:       "A",
		LearningTarget: "recognize",
	},
	{
		ID:             2,
		Term:           "able",
		Phonetic:       "/ˈeɪbl/",
		PartOfSpeech:   "adj",
		Translation:    "能够的；有能力的",
		Examples:       []string{"be able to do sth.", "She is able to speak three languages."},
		Theme:          "general",
		Priority:       "B",
		LearningTarget: "recognize",
	},
	{
		ID:             3,
		Term:           "about",
		Phonetic:       "/əˈbaʊt/",
		PartOfSpeech:   "adv & prep",
		Translation:    "关于；大约",
		Examples:       []string{"a book about history", "It's about three o'clock."},
		Theme:          "general",
		Priority:       "A",
		LearningTarget: "recognize",
	},
	{
		ID:             4,
		Term:           "accept",
		Phonetic:       "/əkˈsept/",
		PartOfSpeech:   "v",
		Translation:    "接受；同意",
		Examples:       []string{"I accept your apology."},
		Theme:          "general",
		Priority:       "B",
		LearningTarget: "recognize",
	},
	{
		ID:             5,
		Term:           "adventure",
		Phonetic:       "/ədˈventʃər/",
		PartOfSpeech:   "n",
		Translation:    "冒险；奇遇",
		Examples:       []string{"We had many adventures on our trip."},
		Theme:          "travel",
		Priority:       "C",
		LearningTarget: "recognize",
	},
}
