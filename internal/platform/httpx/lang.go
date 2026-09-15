package httpx

import (
	"net/http"
	"strings"
)

// Bangla copy for every stable error code.
//
// The code is the contract and never changes; the message is for a human. Most of
// the audience reads Bangla more comfortably than English, and an error is exactly
// where plain language matters most -- so this is product surface, not decoration.
//
// Keep it in step with the English strings at the call sites. An unlisted code simply
// falls back to English rather than showing a placeholder.
var bangla = map[string]string{
	"missing_api_key":         "আপনার API কী Bearer টোকেন হিসেবে পাঠান।",
	"session_required":        "এটি বদলাতে সাইন ইন করুন। API কী দিয়ে হবে না।",
	"not_permitted":           "দলে কে আছে তা কেবল মালিক বা অ্যাডমিন বদলাতে পারেন।",
	"invalid_role":            "অ্যাডমিন বা সদস্য বেছে নিন। নতুন মালিক কেবল মালিকই করতে পারেন।",
	"already_a_member":        "এই ব্যক্তি আগে থেকেই দলে আছেন।",
	"last_owner":              "আগে অন্য কাউকে মালিক করুন। মালিক ছাড়া অ্যাকাউন্ট রাখা যায় না।",
	"cannot_remove_self":      "নিজেকে সরানো যায় না। অন্য কোনো মালিককে বলুন।",
	"invite_not_found":        "আমন্ত্রণটির মেয়াদ শেষ বা এটি ব্যবহার হয়ে গেছে। নতুন একটি চান।",
	"invalid_image":           "PNG, JPEG বা WebP ছবি পাঠান।",
	"unknown_profile":         "এই নামে কোনো এনকোডিং প্রিসেট নেই।",
	"unknown_provider":        "এই সেবা থেকে ভিডিও আনা যায় না। কিছু হোস্ট মূল ফাইল দেয় না।",
	"invalid_edit":            "এই এডিট দিয়ে ভিডিও বানানো যায় না।",
	"no_logo":                 "এই অ্যাকাউন্টে কোনো লোগো নেই। আগে লোগো আপলোড করুন।",
	"not_ready":               "ভিডিওটি এখনও প্রস্তুত হয়নি। প্রস্তুত হলে এডিট করা যাবে।",
	"image_too_large":         "ছবিটি ১ MB এর বেশি। ছোট একটি পাঠান।",
	"invalid_state":           "এই মুহূর্তে এটি করা যাবে না।",
	"invalid_api_key":         "এই API কী টি সঠিক নয়।",
	"asset_not_found":         "এই ভিডিওটি খুঁজে পাওয়া যায়নি।",
	"not_found":               "ফাইলটি খুঁজে পাওয়া যায়নি।",
	"invalid_request":         "JSON বডিতে \"url\" ফিল্ডটি পাঠান।",
	"missing_url":             "ভিডিওর লিংকটি দিন।",
	"invalid_url":             "এটি সঠিক http বা https লিংক নয়।",
	"unknown_event":           "এই নামে কোনো ইভেন্ট আমরা পাঠাই না।",
	"invalid_date":            "RFC 3339 ফরম্যাটে তারিখ দিন, যেমন 2026-09-01T00:00:00Z।",
	"invalid_range":           "সময়সীমার শেষ তারিখ শুরুর পরে হতে হবে।",
	"quota_exceeded":          "আপনার সীমা শেষ হয়ে গেছে। চলমান কাজ শেষ হলে আবার চেষ্টা করুন।",
	"playback_not_authorized": "এই লিংকের মেয়াদ শেষ হয়ে গেছে বা এটি বৈধ নয়।",
	"range_not_satisfiable":   "ফাইলের এই অংশটি নেই।",
	"storage_unavailable":     "ভিডিওটি সাময়িকভাবে পাওয়া যাচ্ছে না। একটু পরে চেষ্টা করুন।",
	"database_unavailable":    "সেবাটি সাময়িকভাবে বন্ধ আছে। একটু পরে চেষ্টা করুন।",
	"internal_error":          "আমাদের দিকে কিছু একটা সমস্যা হয়েছে।",
	"source_unreadable":       "ভিডিও ফাইলটি পড়া যায়নি।",
	"no_video_stream":         "ফাইলটিতে কোনো ভিডিও পাওয়া যায়নি।",
	"source_too_large":        "ভিডিও ফাইলটি অনুমোদিত সীমার চেয়ে বড়।",
	"source_url_not_allowed":  "এই লিংক থেকে ভিডিও নেওয়া যাবে না।",
	"source_unreachable":      "লিংকটি থেকে ভিডিও নামানো যায়নি।",
	"encode_failed":           "ভিডিওটি প্রসেস করা যায়নি।",
	"processing_failed":       "ভিডিওটি প্রসেস করা যায়নি।",
	"stitch_failed":           "ভিডিওটি প্রসেস করা যায়নি।",
	"package_failed":          "ভিডিওটি প্রসেস করা যায়নি।",
	"invalid_ladder_profile":  "আপনার অ্যাকাউন্টের ভিডিও সেটিংসে সমস্যা আছে।",
	"stream_not_found":        "এই লাইভ স্ট্রিমটি খুঁজে পাওয়া যায়নি।",
	"stream_busy":             "এই স্ট্রিমটি আগে থেকেই এনকোডারের অপেক্ষায় আছে।",
	"no_ingest_port":          "এখন নতুন লাইভ স্ট্রিম নেওয়া যাচ্ছে না। একটু পরে চেষ্টা করুন।",
	"invalid_protocol":        "srt অথবা rtmp বেছে নিন।",
	"ingest_failed":           "এনকোডার থেকে ভিডিও নেওয়া যায়নি।",
}

// ErrorCodes lists every code with Bangla copy, so a test can assert that the set
// stays in step with what the handlers actually emit.
func ErrorCodes() []string {
	out := make([]string, 0, len(bangla))
	for c := range bangla {
		out = append(out, c)
	}
	return out
}

// Localise picks the message language from Accept-Language, defaulting to English.
func Localise(r *http.Request, code, english string) string {
	if prefersBangla(r.Header.Get("Accept-Language")) {
		if bn, ok := bangla[code]; ok {
			return bn
		}
	}
	return english
}

// prefersBangla does a deliberately simple check: full RFC 4647 negotiation buys
// nothing when there are two languages and one of them is the fallback.
func prefersBangla(header string) bool {
	for _, part := range strings.Split(header, ",") {
		tag := strings.ToLower(strings.TrimSpace(strings.SplitN(part, ";", 2)[0]))
		switch {
		case tag == "bn" || strings.HasPrefix(tag, "bn-"):
			return true
		case tag == "en" || strings.HasPrefix(tag, "en-"):
			return false
		}
	}
	return false
}

// ErrorFor writes a failure in the language the caller asked for.
func ErrorFor(w http.ResponseWriter, r *http.Request, status int, code, english string) {
	Error(w, status, code, Localise(r, code, english))
}
