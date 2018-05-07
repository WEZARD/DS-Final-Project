<!DOCTYPE html>
<html>
<head>
    <title></title>
</head>
<div class="row">
    <div class="container">
        <h1>Login Page</h1>
        <form name="login_form" method="post" action="/login" onSubmit="return InputCheck(this)" autocomplete="off">
            <label for="username">Username</label>
            <input class="form-control" style="width:20%" type="text" id="username" name="username" value="">
            <label for="password">Password</label>
            <input class="form-control" style="width:20%" type="password" id="password" name="password" value=""><br />
            <button type="submit" class="btn btn-default">Login</button>   
        </form><br />
        <ul class="fr hd-bar">
            <li>Customer Service：<span>9292888888</span></li>
            <li class="active"><a href="/login">[Sign In]</a></li>
            <li><a href="/register">[Sign Up]</a></li>
        </ul>
    </div>
</div>

<script language=JavaScript>
<!--
function InputCheck(login_form) { 
    if (login_form.username.value == "") {
        alert("Please enter your username!");  
        login_form.username.focus();  
        return (false);
    }

    if (login_form.password.value == "") {
        alert("Please enter your password!");  
        login_form.password.focus();  
        return (false);  
    } 
} 
//-->
</script>

<!-- Latest compiled and minified CSS -->
<link rel="stylesheet" href="https://maxcdn.bootstrapcdn.com/bootstrap/3.3.7/css/bootstrap.min.css">