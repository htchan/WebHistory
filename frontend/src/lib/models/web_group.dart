import 'web.dart';

class WebGroup {
  final List<Web> webs;

  WebGroup(this.webs) {
    webs.sort( (a, b) => a.updateTime.compareTo(b.updateTime));
  }

  Web get latestWeb {
    return webs.last;
  }
}