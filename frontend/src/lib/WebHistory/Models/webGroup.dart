import './web.dart';

class WebGroup {
  final List<Web> webs;

  WebGroup(this.webs);

  Web get latestWeb {
    Web target = webs[0];
    for (var i = 1; i < webs.length; i++) {
      if (webs[i].updateTime.isAfter(target.updateTime)) {
        target = webs[i];
      }
    }
    return target;
  }
}