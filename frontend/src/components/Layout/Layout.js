import React from 'react';
import NavBar from '../NavBar/NavBar';

const Layout = ({ children, logout }) => {
  return (
    <>
      <NavBar logout={logout} />
      <div className="content">
        {children}
      </div>
    </>
  );
};

export default Layout;
