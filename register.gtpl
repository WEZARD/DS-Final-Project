<!DOCTYPE html>
<html>
<head>
	<title></title>

<script language=JavaScript>
<!--

function InputCheck(register_form) { 
  if (register_form.username.value == "") {
    alert("Please enter your username!");  
    register_form.username.focus();  
    return (false);
  }

  if (register_form.password.value == "") {
    alert("Please enter your password!");  
    register_form.password.focus();  
    return (false);  
  } 
} 

//-->
</script>


</head>
<div class="row">
   <div class="container">
<h1>Register Page</h1>

<form name="register_form" method="post" action="/register" onSubmit="return InputCheck(this)" autocomplete="off">
    <label for="username">Username</label>
    <input class="form-control" type="text" id="username" name="username" value="">
    <label for="password">Password</label>
    <input class="form-control" type="password" id="password" name="password" value="">
    <button type="submit">Register</button>
    
</form>
<ul class="fr hd-bar">
    <li>Customer Service：<span>9292888888</span></li>
    <li class="active"><a href="/login">[Sign In]</a></li>
    <li><a href="/register">[Sign Up]</a></li>
</ul>
</div>
</div>

<!-- Latest compiled and minified CSS -->
<link rel="stylesheet" href="https://maxcdn.bootstrapcdn.com/bootstrap/3.3.7/css/bootstrap.min.css">