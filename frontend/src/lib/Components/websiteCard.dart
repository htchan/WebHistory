import 'dart:convert';
import 'dart:js';
// ignore: file_names
import 'package:flutter/material.dart';
// ignore: import_of_legacy_library_into_null_safe
import 'package:http/http.dart' as http;
import 'package:url_launcher/url_launcher.dart';
import 'package:flutter_slidable/flutter_slidable.dart';

class WebsiteCard extends StatelessWidget {
  final String url;
  // final List<Map> websiteGroup;
  final Function updateListCallBack, openDetailsPage;
  final Function? showChangeGroupDialog;
  final Map website;
  final bool isEdit;
  final String token;
  WebsiteCard(this.url, this.website, this.token, this.updateListCallBack, this.openDetailsPage, {this.isEdit = false, this.showChangeGroupDialog});

  void openURL() async {
    final String apiUrl = '$url/websites/${website['uuid']}/refresh';
    http.put(
      Uri.parse(apiUrl),
      body: <String, String>{
        'url': website['url']??"",
      },
      headers: {"Authorization": token}
    )
    .then( (response) => updateListCallBack() );
    // TODO: if it is not available to launch, it have to give a pop up
    await canLaunch(website['url']??"")? await launch(website['url']??"") : "";
  }

  void removeCard() {
    final String apiUrl = '$url/websites/${website['uuid']}/';
    // final String targetIdentifier = website['url'] == '' ? 'unknown' : (website['url']??"unknown");
    http.delete(
      Uri.parse('$apiUrl'),
      headers: {"Authorization": token}
    )
    .then((response) {
      if (response.statusCode >= 200 && response.statusCode < 300) {
        updateListCallBack();
      }
    });
  }

  Text renderSubTitleText(DateTime accessTime, DateTime updateTime) {
    return Text(
      (website['url']??"") + '\n' +
      'Update Time: ' + updateTime.toLocal().toString() + '\n' +
      'Access Time: ' + accessTime.toLocal().toString()
    );
  }

  Widget renderStatusIcon(bool updated)  {
    return updated ? 
      Container( color: Colors.green, child: const Icon(Icons.check_circle, color: Colors.white) ) : 
      Container( color: Colors.red, child: const Icon(Icons.remove_circle, color: Colors.white) );
  }

  Widget renderDeleteAction() {
    return IconSlideAction(
      caption: 'Delete',
      color: Colors.red,
      icon: Icons.delete,
      onTap: removeCard
    );
  }
  Widget renderDetailsAction() {
    return IconSlideAction(
      caption: "Details",
      color: Colors.blue,
      icon: Icons.info,
      onTap: () => openDetailsPage(website["group_name"])
    );
  }
  Widget renderChangeGroupAction() {
    return IconSlideAction(
      caption: "Change Group",
      color: Colors.yellow,
      icon: Icons.edit,
      onTap: () {
        print("working");
        // show a dialog for input / select new group name (default group is user )
        showChangeGroupDialog!(website);
        // update the page
        updateListCallBack();
      }
    );
  }

  @override
  Widget build(BuildContext context) {
    var accessTime = DateTime.parse(website['access_time']??"20121225T0000");
    var updateTime = DateTime.parse(website['update_time']??"20121225T0000");
    return Slidable(
      actionPane: SlidableDrawerActionPane(),
      actionExtentRatio:0.2,
      child: GestureDetector(
        onTap: openURL,
        child:ListTile(
          leading: renderStatusIcon(
            accessTime.millisecondsSinceEpoch < updateTime.millisecondsSinceEpoch
          ),
          title: Text(website['title']??"Load title fail"),
          subtitle: renderSubTitleText(accessTime, updateTime),
        ),
      ),
      actions: [
        (isEdit) ? renderChangeGroupAction() : renderDetailsAction(),
        renderDeleteAction()
      ],
      secondaryActions: [
        (isEdit) ? renderChangeGroupAction() : renderDetailsAction(),
        renderDeleteAction()
      ],
    );
  }
}