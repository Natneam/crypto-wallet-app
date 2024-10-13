import React from "react";
import "./Navbar.css";

function Navbar({ logout }) {
  return (
    <nav className="navbar">
      <h2>Crypto Wallet App</h2>
      <button onClick={() => logout()} className="logout-button">Logout</button>
    </nav>
  );
}

export default Navbar;
