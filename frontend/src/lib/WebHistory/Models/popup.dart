class Popup {
  final String content;
  Popup.from(Map<String, String> map):
    this.content = map['message']??map['error']??'unknown message';
}