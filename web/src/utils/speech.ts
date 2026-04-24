/**
 * Web Speech API 封装（浏览器原生，免后端）
 * - 使用 SpeechSynthesis 进行单词朗读
 * - 统一语言与音高，避免组件重复设置
 */
const DEFAULT_LANG = 'en-US'

export function isSpeechSupported(): boolean {
  return typeof window !== 'undefined' && 'speechSynthesis' in window
}

export function speak(text: string, lang: string = DEFAULT_LANG): void {
  if (!isSpeechSupported() || !text) return
  const synth = window.speechSynthesis
  // 打断当前朗读，避免快速切换时累积
  synth.cancel()
  const utter = new SpeechSynthesisUtterance(text)
  utter.lang = lang
  utter.rate = 0.95
  utter.pitch = 1
  synth.speak(utter)
}

export function stopSpeak(): void {
  if (!isSpeechSupported()) return
  window.speechSynthesis.cancel()
}
