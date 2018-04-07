<!DOCTYPE html>
<html>
<head>
  <title></title>
</head>
<div class="row">
  <div class="container">
    <div style="width:100%; height:100px">
      <div style="float:left; width:50%">
	<h1>Home Page</h1>
      </div>
      <div style="float:right; width:50%; height:100%; text-align:center; margin-top:20px">
	<form action="/cancelaccount" method="post" name="cancelaccount_form">
	  <input type="submit" value="Cancle Account">
	</form>
      </div>
    </div>
    <div id＝"main">
       <div id="follower" style="float:left; width:30%">
	 <div style="float:left; width:80px">Followers</div>
	 <div style="margin-left: 80px">
	   <form action="/addfollowing" method="post" name="addfollow_form">
	     <input type="submit" value="Add follow">
	   </form>
	 </div><br />
	 <p>user one</p>
	 <p>user two</p>
       </div>
       <div id="message" style="margin-left: 30%">
	 <p>Message</p>
	 <p>message one</p>
	 <p>message two</p>
       </div>
     </div>
   </div>
</div>
<body>
    <div style="color:green; text-align:center; position:absolute; bottom:0; width:100%;" class="footer">
      <form action="post.php" method="post">
	<p>Share with your friends:</p>
	<input type="text" name="postcontent" style="width:80%; height:100px"><br>
	<p></p><p></p><input type="submit" value="post"></p>
      </form>
    </div>
  </body>


<!-- Latest compiled and minified CSS -->
<link rel="stylesheet" href="https://maxcdn.bootstrapcdn.com/bootstrap/3.3.7/css/bootstrap.min.css">
