<!DOCTYPE html>
<html>
<head>
  <title></title>
<script language=JavaScript>
<!--

function InputCheck(login_form) { 
  if (addfollow_form.username.value == "") {
    alert("Please enter username!");  
    login_form.username.focus();  
    return (false);
  }
} 

//-->
</script>
</head>
<div class="row">
   <div class="container">
<h1>Add Follow</h1>

<form name="addfollow_form" method="post" action="/addfollow" onSubmit="return InputCheck(this)" autocomplete="off">
    <label for="username">Username </label>
    <input class="form-control" type="text" id="username" name="username" value="">
    <button type="submit">Search</button>
    
</form>
</div>
</div>

<!-- Latest compiled and minified CSS -->
<link rel="stylesheet" href="https://maxcdn.bootstrapcdn.com/bootstrap/3.3.7/css/bootstrap.min.css">
